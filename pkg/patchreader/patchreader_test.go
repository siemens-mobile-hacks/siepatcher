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
	}

	for _, tc := range testCases {
		p, err := FromFile(testFileFullPath(tc.fileName))
		if (err != nil) != tc.wantError {
			t.Fatalf("Test failure: %t (%v), want %t", err != nil, err, tc.wantError)
		}
		if p.NumChunks() != tc.NumChunks {
			t.Fatalf("Got %d chunks in patch, want %d", p.NumChunks(), tc.NumChunks)
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
