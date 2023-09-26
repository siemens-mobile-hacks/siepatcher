package patchreader

import (
	"bytes"
	"path/filepath"
	"testing"
)

func testFileFullPath(baseName string) string {
	return filepath.Join("..", "..", "testdata", baseName)
}

func TestLoadPlainPatch(t *testing.T) {
	testCases := []struct {
		fileName  string
		NumChunks int
		wantError bool
	}{
		{
			fileName:  "plainbody.vkp",
			NumChunks: 1,
			wantError: false,
		},
		{
			fileName:  "onebigchunk.vkp",
			NumChunks: 1,
			wantError: false,
		},
		{
			fileName:  "pragma_old_equal_ff.vkp",
			NumChunks: 2,
			wantError: false,
		},
		{
			fileName:  "addr_offset.vkp",
			NumChunks: 3,
			wantError: false,
		},
		{
			fileName:  "addr_offset_1.vkp",
			NumChunks: 2,
			wantError: false,
		},
		{
			fileName:  "comma_separated_data.vkp",
			NumChunks: 2,
			wantError: false,
		},
		{
			fileName:  "multiline_comments.vkp",
			NumChunks: 2,
			wantError: false,
		},
		{
			fileName:  "ints_in_data.vkp",
			NumChunks: 1,
			wantError: false,
		},
		{
			fileName:  "more_old_than_new.vkp",
			NumChunks: 1,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		p, err := FromFile(testFileFullPath(tc.fileName))
		if (err != nil) != tc.wantError {
			t.Fatalf("Test %q: %t (%v), want %t", tc.fileName, err != nil, err, tc.wantError)
		}
		if p.NumChunks() != tc.NumChunks {
			t.Fatalf("Test %q: Got %d chunks in patch, want %d", tc.fileName, p.NumChunks(), tc.NumChunks)
		}
	}

}

func TestParsePragma(t *testing.T) {
	testCases := []struct {
		pragmaStr string
		wantError bool
	}{
		{
			pragmaStr: "#pragma enable old_equal_ff",
			wantError: false,
		},
		{
			pragmaStr: "#pragma disable old_equal_ff",
			wantError: false,
		},
		{
			pragmaStr: "#pragma disalbe old_equal_ff",
			wantError: true,
		},
	}

	for _, tc := range testCases {
		var settings chunkSettings
		err := parsePragma(&settings, tc.pragmaStr)
		if (err != nil) != tc.wantError {
			t.Fatalf("Test string %q: %t (%v), want %t", tc.pragmaStr, err != nil, err, tc.wantError)
		}

	}
}

func TestParseDataField(t *testing.T) {
	testCases := []struct {
		descr           string
		dataFieldString string
		wantBytes       []byte
		wantError       bool
	}{
		{
			descr:           "A normal hexadecimal string",
			dataFieldString: "AA00BB",
			wantBytes:       []byte{0xAA, 0x00, 0xBB},
			wantError:       false,
		},
		{
			dataFieldString: "0xA0345678",
			wantBytes:       []byte{0xA0, 0x34, 0x56, 0x78},
			wantError:       false,
		},
		{
			dataFieldString: "A0,B1,C2D3,E4,F5",
			wantBytes:       []byte{0xA0, 0xB1, 0xC2, 0xD3, 0xE4, 0xF5},
			wantError:       false,
		},
		{
			descr:           "Decimal numbers get parsed",
			dataFieldString: "0i28",
			wantBytes:       []byte{28},
			wantError:       false,
		},
		{
			descr:           "Decimal numbers get parsed #2",
			dataFieldString: "0i255",
			wantBytes:       []byte{0xFF},
			wantError:       false,
		},
		{
			descr:           "Decimal numbers get parsed and padded",
			dataFieldString: "0i00255",
			wantBytes:       []byte{0xFF, 0x00},
			wantError:       false,
		},

		{
			descr:           "A number doesn't fit in the space defined by its length",
			dataFieldString: "0i256",
			wantBytes:       []byte{},
			wantError:       true,
		},
		{
			descr:           "A number padded to len=5 does fit",
			dataFieldString: "0i00256",
			wantBytes:       []byte{0, 1},
			wantError:       false,
		},
		{
			descr:           "Negative numbers get parsed",
			dataFieldString: "0i-1",
			wantBytes:       []byte{0xFF},
			wantError:       false,
		},

		{
			wantError: false,
		},
	}

	for _, tc := range testCases {
		out, err := parseDataField(tc.dataFieldString)
		if (err != nil) != tc.wantError {
			t.Fatalf("Test %q: test string %q: failure=%t (%v), want %t", tc.descr, tc.dataFieldString, err != nil, err, tc.wantError)
		}
		if !bytes.Equal(out, tc.wantBytes) {
			t.Fatalf("Test %q: got bytes %v, want %v", tc.descr, out, tc.wantBytes)
		}
	}
}
