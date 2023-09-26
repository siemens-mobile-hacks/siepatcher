package patchreader

import (
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
