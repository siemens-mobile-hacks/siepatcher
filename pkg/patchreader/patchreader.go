package patchreader

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	CommentMarker = ';'
	PragmaMarker  = "#pragma"
)

type Chunk struct {
	BaseAddr int64
	OldData  []byte
	NewData  []byte
}

type PatchReader struct {
	txt    string
	chunks []Chunk
}

func FromFile(path string) (*PatchReader, error) {
	p := &PatchReader{}
	txt, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p.txt = string(txt)

	if err := p.parse(); err != nil {
		return nil, err
	}
	return p, nil
}

func (pr *PatchReader) String() string {
	return fmt.Sprintf("Patch with %d chunks", pr.NumChunks())
}

func (pr *PatchReader) NumChunks() int {
	return len(pr.chunks)
}

////////////////////////////////////////////////////////////////////////////
// VKP file format: http://www.vi-soft.com.ua/siemens/vkp_file_format.txt //
////////////////////////////////////////////////////////////////////////////

func parseDataField(df string) ([]byte, error) {
	byteData, err := hex.DecodeString(df)
	if err != nil {
		return nil, err
	}

	return byteData, nil
}

type chunkSettings struct {
	isOldEqualFF bool
	addrOffset   int64
}

// parsePragma recognizes #pragma statements which can change
// the way the patch is being applied.
// The string looks like:
// #pragma enable old_equal_ff
// Supported pragmas:
// * old_equal_ff
// Unsupported pragmas:
// * undo
// * warn_if_new_exist_on_apply
// * warn_no_old_on_apply
func parsePragma(currentSettings *chunkSettings, pragmaStr string) error {
	pragmaPos := strings.Index(pragmaStr, PragmaMarker)
	if pragmaPos == -1 {
		return fmt.Errorf("cannot find #pragma string")
	}
	pragmaBody := pragmaStr[pragmaPos+len(PragmaMarker)+1:]
	pragma := strings.Split(pragmaBody, " ")
	if len(pragma) != 2 {
		return fmt.Errorf("cannot recognize pragma in string %q", pragmaBody)
	}
	if pragma[0] != "enable" && pragma[0] != "disable" {
		return fmt.Errorf("pragma in string %q is neither enabled nor disabled", pragmaBody)
	}
	pragmaEnable := false
	if pragma[0] == "enable" {
		pragmaEnable = true
	}
	switch pragma[1] {
	case "old_equal_ff":
		currentSettings.isOldEqualFF = pragmaEnable
	default:
		return fmt.Errorf("unrecognized pragma %q", pragma[1])
	}
	return nil
}

// We get a string like +0x345 here.
func parseAddrOffset(currentSettings *chunkSettings, offsetStr string) error {

	sign := offsetStr[0]
	offStr := offsetStr[1:]
	var intValue int64
	var err error
	offStr = strings.TrimPrefix(offStr, "0x") // If there is 0x before the address -- kill it.
	intValue, err = strconv.ParseInt(offStr, 16, 64)
	if err != nil {
		return fmt.Errorf("cannot parse %q as hex string: %v", offStr, err)
	}
	if sign == '-' {
		intValue = -intValue
	}
	currentSettings.addrOffset = intValue

	return nil
}

func (pr *PatchReader) parse() error {
	scanner := bufio.NewScanner(strings.NewReader(pr.txt))

	lineNum := 0
	var currentAddr int64 = 0
	var currentSettings chunkSettings

	for scanner.Scan() {
		lineNum++
		patchLine := scanner.Text()
		log.Printf("Read line: %q", patchLine)

		commentPos := strings.Index(patchLine, ";")
		if commentPos != -1 {
			patchLine = patchLine[:commentPos]
		}
		patchLine = strings.TrimSpace(patchLine)
		log.Printf("Line w/o comments: %q", patchLine)

		// If there is nothing left in the string -- ignore it.
		if len(patchLine) == 0 {
			continue
		}

		if strings.HasPrefix(patchLine, PragmaMarker) {
			if err := parsePragma(&currentSettings, patchLine); err != nil {
				return fmt.Errorf("cannot parse pragma on line %d: %v", lineNum, err)
			}
			continue
		}

		if patchLine[0] == '+' || patchLine[0] == '-' {
			if err := parseAddrOffset(&currentSettings, patchLine); err != nil {
				return fmt.Errorf("cannot parse address offset on line %d: %v", lineNum, err)
			}
			continue
		}

		addrPos := strings.Index(patchLine, ":")
		if addrPos == -1 {
			return fmt.Errorf("no address info found in line %d", lineNum)
		}
		addrHex := patchLine[:addrPos]
		addr, err := strconv.ParseInt(addrHex, 16, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %q to int64: %v", addrHex, err)
		}
		log.Printf("Address: %X", addr)

		dataInfo := strings.TrimSpace(patchLine[addrPos+1:])
		dataFields := strings.Split(dataInfo, " ")

		var oldData []byte
		var newDataStr string
		if !currentSettings.isOldEqualFF {
			if len(dataFields) != 2 {
				return fmt.Errorf("cannot split string %q into data information", dataInfo)
			}
			var err error
			oldData, err = parseDataField(dataFields[0])
			if err != nil {
				return fmt.Errorf("cannot parse old data: %v", err)
			}
			newDataStr = dataFields[1]
		} else {
			if len(dataFields) != 1 {
				return fmt.Errorf("cannot split string %q into data information (old_equal_ff enabled)", dataInfo)
			}
			newDataStr = dataFields[0]
		}
		newData, err := parseDataField(newDataStr)
		if err != nil {
			return fmt.Errorf("cannot parse new data: %v", err)
		}

		if currentSettings.isOldEqualFF {
			oldData = make([]byte, len(newData))
			for i := 0; i < len(newData); i++ {
				oldData[i] = 0xFF
			}
		}

		// If old and new data have different lengths -- this is a problem.
		if len(oldData) != len(newData) {
			return fmt.Errorf("old data length (%d) is not equal to new data length (%d)", len(oldData), len(newData))
		}

		// Now, if this line is describing a continuos block of data together with the previous line,
		// just extend the previous line.
		// If this line describes the changes at an address that doesn't follow immediately after the prev line,
		// create a new chunk.
		if currentAddr == addr {
			lastChunk := &pr.chunks[len(pr.chunks)-1]
			lastChunk.OldData = append(lastChunk.OldData, oldData...)
			lastChunk.NewData = append(lastChunk.NewData, newData...)
		} else {
			newChunk := Chunk{}
			newChunk.BaseAddr = addr
			newChunk.OldData = oldData
			newChunk.NewData = newData
			pr.chunks = append(pr.chunks, newChunk)
			currentAddr = addr
		}
		currentAddr += int64(len(newData))
	}
	return nil
}
