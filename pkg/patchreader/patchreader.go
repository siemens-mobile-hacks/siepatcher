package patchreader

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math"
	"math/bits"
	"os"
	"regexp"
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

func (c *Chunk) Size() int64 {
	return int64(len(c.NewData))
}

func (c *Chunk) EndAddr() int64 {
	return c.BaseAddr + c.Size()
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

func (pr *PatchReader) Chunks() []Chunk {
	return pr.chunks
}

////////////////////////////////////////////////////////////////////////////
// VKP file format: http://www.vi-soft.com.ua/siemens/vkp_file_format.txt //
////////////////////////////////////////////////////////////////////////////

func parseDecimalNum(dataBlock string) ([]byte, error) {
	outBuf := make([]byte, 0)

	isSigned := dataBlock[0] == '-'
	numberLen := len(dataBlock)
	if isSigned {
		numberLen -= 1
	}
	intValue, err := strconv.ParseInt(dataBlock, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q as int string: %v", dataBlock, err)
	}
	nBytes := 1
	if numberLen >= 5 {
		nBytes = 2
	}
	if numberLen >= 8 {
		nBytes = 3
	}
	if numberLen >= 10 {
		nBytes = 4
	}
	if numberLen >= 13 {
		nBytes = 5
	}
	if numberLen >= 15 {
		nBytes = 6
	}
	if numberLen >= 17 {
		nBytes = 7
	}
	if numberLen >= 20 {
		nBytes = 8
	}

	reqBits := bits.Len64(uint64(math.Abs(float64(intValue))))
	if isSigned {
		reqBits += 1
	}
	// v_Klay doesn't detect this, but we do.
	if reqBits > nBytes*8 {
		return nil, fmt.Errorf("Need at least %d bytes to represent %d, but have only %d", reqBits, intValue, nBytes*8)
	}

	for i := 0; i < nBytes; i++ {
		var b uint8
		b = byte((intValue >> (i * 8)) & 0xFF)
		outBuf = append(outBuf, b)
	}
	return outBuf, nil
}

func parseDataField(df string) ([]byte, error) {
	outBuf := make([]byte, 0)

	dataBlocks := strings.Split(df, ",")
	for _, dataBlock := range dataBlocks {
		if strings.HasPrefix(dataBlock, "0i") {
			dataBlock = strings.TrimPrefix(dataBlock, "0i") // 0i46 --> 46
			byteData, err := parseDecimalNum(dataBlock)
			if err != nil {
				return nil, err
			}
			outBuf = append(outBuf, byteData...)
			continue
		}
		dataBlock = strings.TrimPrefix(dataBlock, "0x") // 0xA04B1C70 --> A04B1C70
		byteData, err := hex.DecodeString(dataBlock)
		if err != nil {
			return nil, err
		}
		outBuf = append(outBuf, byteData...)
	}
	return outBuf, nil
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

func removeMultilineComments(text string) string {
	r := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return r.ReplaceAllString(text, "")
}

func (pr *PatchReader) parse() error {

	// Remove mult-line comments.
	pr.txt = removeMultilineComments(pr.txt)

	scanner := bufio.NewScanner(strings.NewReader(pr.txt))

	lineNum := 0
	var currentAddr int64 = 0
	var currentSettings chunkSettings

	for scanner.Scan() {
		lineNum++
		patchLine := scanner.Text()

		commentPos := strings.Index(patchLine, ";")
		if commentPos != -1 {
			patchLine = patchLine[:commentPos]
		}
		patchLine = strings.TrimSpace(patchLine)

		// If there is nothing left in the string -- ignore it.
		if len(patchLine) == 0 {
			continue
		}

		if strings.HasPrefix(patchLine, PragmaMarker) {
			if err := parsePragma(&currentSettings, patchLine); err != nil {
				return fmt.Errorf("line %d: cannot parse pragma: %v", lineNum, err)
			}
			continue
		}

		if patchLine[0] == '+' || patchLine[0] == '-' {
			if err := parseAddrOffset(&currentSettings, patchLine); err != nil {
				return fmt.Errorf("line %d: cannot parse address offset: %v", lineNum, err)
			}
			continue
		}

		addrPos := strings.Index(patchLine, ":")
		if addrPos == -1 {
			return fmt.Errorf("line %d: no address info found", lineNum)
		}
		addrHex := strings.TrimPrefix(patchLine[:addrPos], "0x")
		addr, err := strconv.ParseInt(addrHex, 16, 64)
		if err != nil {
			return fmt.Errorf("line %d: cannot convert address %q to int64: %v", lineNum, addrHex, err)
		}

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
		if currentAddr == addr && len(pr.chunks) != 0 {
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
