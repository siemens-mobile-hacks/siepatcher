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

func parseDataField(df string) ([]byte, error) {
	byteData, err := hex.DecodeString(df)
	if err != nil {
		return nil, err
	}

	return byteData, nil
}

func (pr *PatchReader) parse() error {
	scanner := bufio.NewScanner(strings.NewReader(pr.txt))
	lineNum := 0
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

		newChunk := Chunk{}

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
		newChunk.BaseAddr = addr

		dataInfo := strings.TrimSpace(patchLine[addrPos+1:])
		dataFields := strings.Split(dataInfo, " ")
		if len(dataFields) != 2 {
			return fmt.Errorf("cannot split string %q into data information", dataInfo)
		}
		oldData, err := parseDataField(dataFields[0])
		if err != nil {
			return fmt.Errorf("cannot parse old data: %v", err)
		}
		newData, err := parseDataField(dataFields[1])
		if err != nil {
			return fmt.Errorf("cannot parse new data: %v", err)
		}
		newChunk.OldData = oldData
		newChunk.NewData = newData

		pr.chunks = append(pr.chunks, newChunk)
	}
	return nil
}
