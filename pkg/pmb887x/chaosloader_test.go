package pmb887x

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestParseChaosInfo(t *testing.T) {
	testCases := []struct {
		descr            string
		chaosReply       string
		wantFlashSize    int64
		wantFlashRegions int
	}{
		{
			descr:            "EL71, 64MB, one flash region",
			chaosReply:       "454C37310000000000000000000000005349454D454E53000000000000000000585858585858585858585858585858008F77473E07433B6A6AA7A8BC4217BD5A000000A0A975DC16000300000000000020001988010A0201FF000004FFFFFFFFFFFFFFFFFFFFFFFF000000000000000000000000000000000000000000000000",
			wantFlashSize:    64 * 1024 * 1024,
			wantFlashRegions: 1,
		},
		{
			descr:            "C81, 64MB, four flash regions",
			chaosReply:       "433831000000000000000000000000005349454D454E5300000000000000000058585858585858585858585858585800664C544260E5CC2931FBF4799D65BE27000000A003C25490000300000000000089000D8802060004FE0000020300800003008000FE0000025052493133A6000000000000000000000000000000000000",
			wantFlashSize:    64 * 1024 * 1024,
			wantFlashRegions: 4,
		},
	}

	for _, tc := range testCases {
		byteData, err := hex.DecodeString(tc.chaosReply)
		if err != nil {
			t.Fatalf("Test %q: Cannot prepare data for Chaos reply: %v", tc.descr, err)
		}
		info, err := ParseChaosInfo(bytes.NewBuffer(byteData))
		if err != nil {
			t.Fatalf("Test %q: Cannot parse Chaos reply: %v", tc.descr, err)
		}
		if info.BlockMap.TotalSize() != tc.wantFlashSize {
			t.Fatalf("Test %q: Unexpected flash size: got %d, want %d.\nBlockmap: %s", tc.descr, info.BlockMap.TotalSize(), tc.wantFlashSize, info.BlockMap)
		}
		if info.BlockMap.NumOfRegions() != tc.wantFlashRegions {
			t.Fatalf("Test %q: Unexpected number of regions: got %d, want %d.\nBlockmap: %s", tc.descr, info.BlockMap.NumOfRegions(), tc.wantFlashRegions, info.BlockMap)
		}
	}
}
