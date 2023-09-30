package blockman

// BlockmapForC81 returns a block map that is used in a real C81.
func BlockmapForC81() Blockman {
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

	return bm
}
