package blockman

import "testing"

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
			blockSize: 0x1000000,
		},
		{
			desc:      "Normal address within flash #2",
			addr:      0xA1000000,
			wantError: false,
			blockAddr: 0xA1000000,
			blockSize: 0x1000000,
		},
		{
			desc:      "Normal address within flash #3",
			addr:      0xA1000000 - 1,
			wantError: false,
			blockAddr: 0xA0000000,
			blockSize: 0x1000000,
		},
		{
			desc:      "Address not in flash",
			addr:      0xA8001000,
			wantError: true,
		},
	}

	bm := New(0xA0000000)
	bm.AddRegion(0x1000000, 8)
	bm.AddRegion(0x0800000, 16)

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
