package blockman

import (
	"testing"
)

func TestParamsForAddr(t *testing.T) {
	testCases := []struct {
		desc      string
		addr      int64
		wantError bool
		blockAddr int64
		blockSize int64
	}{
		{
			desc:      "Normal address within flash #1",
			addr:      0xA0001000,
			wantError: false,
			blockAddr: 0xA0000000,
			blockSize: 0x20000,
		},
		{
			desc:      "Normal address within flash #2",
			addr:      0xA0020001,
			wantError: false,
			blockAddr: 0xA0020000,
			blockSize: 0x20000,
		},
		{
			desc:      "Normal address within flash, in the second region",
			addr:      0xA1FE0001,
			wantError: false,
			blockAddr: 0xA1FE0000,
			blockSize: 0x8000,
		},
		{
			desc:      "Address not in flash",
			addr:      0xA8001000,
			wantError: true,
		},
	}

	/*
		From the real C81:
		BM: Total size: 64 MB
		4 regions, start addr 0xA0000000, end addr 0xA3FFFFFF
			Region #0: [A0000000, A1FDFFFF] 255 blocks, size of each 0x20000
			Region #1: [A1FE0000, A001FFFF] 4 blocks, size of each 0x8000
			Region #2: [A2000000, A001FFFF] 4 blocks, size of each 0x8000
			Region #3: [A2020000, A1FDFFFF] 255 blocks, size of each 0x20000
	*/

	bm := New(0xA0000000)
	bm.AddRegion(0x20000, 255)
	bm.AddRegion(0x8000, 4)
	bm.AddRegion(0x8000, 4)
	bm.AddRegion(0x20000, 255)

	for _, tc := range testCases {
		blockAddr, blockSize, err := bm.ParamsForAddr(tc.addr)
		if (err != nil) != tc.wantError {
			t.Fatalf("Test %q: failed = %t (%v), want %t", tc.desc, err != nil, err, tc.wantError)
		}

		// If there was an error, doesn't make sense to compare other values.
		if err != nil {
			continue
		}

		if blockAddr != tc.blockAddr {
			t.Errorf("Test %q: got blockAddr = %X, want %X", tc.desc, blockAddr, tc.blockAddr)
		}
		if blockSize != tc.blockSize {
			t.Errorf("Test %q: got blockSize = %X, want %X", tc.desc, blockSize, tc.blockSize)
		}
	}
}
