package blockman

import "fmt"

type BlockRegionInfo struct {
	baseAddr   int64
	endAddr    int64
	blockSize  int64
	blockCount int
}

type Blockman struct {
	baseAddr     int64
	endAddr      int64
	totalSize    int64
	blockRegions []BlockRegionInfo
}

func New(baseAddr int64) Blockman {
	return Blockman{
		baseAddr:     baseAddr,
		endAddr:      baseAddr,
		totalSize:    0,
		blockRegions: []BlockRegionInfo{},
	}
}

func (b *Blockman) AddRegion(blockSize int64, blockCount int) error {
	var regionSize int64 = blockSize * int64(blockCount)
	blockRegion := BlockRegionInfo{
		baseAddr:   b.baseAddr,
		endAddr:    b.baseAddr + regionSize,
		blockSize:  blockSize,
		blockCount: blockCount,
	}
	b.blockRegions = append(b.blockRegions, blockRegion)
	b.endAddr += int64(regionSize)

	return nil
}

func (b *Blockman) ParamsForAddr(addr int64) (baseAddr, size int64, err error) {
	if addr < b.baseAddr || addr >= b.endAddr {
		return -1, -1, fmt.Errorf("addr 0x%X is out of bounds [0x%X, 0x%X)", addr, b.baseAddr, b.endAddr)
	}
	for i := 0; i < len(b.blockRegions); i++ {
		region := b.blockRegions[i]
		if addr < region.baseAddr || addr >= region.endAddr {
			continue
		}
		for blockNo := 0; blockNo < region.blockCount; blockNo++ {
			blockStartAddr := region.baseAddr + region.blockSize*int64(blockNo)
			blockEndAddr := blockStartAddr + region.blockSize
			if addr >= blockStartAddr && addr < blockEndAddr {
				return blockStartAddr, region.blockSize, nil
			}
		}
	}
	return -1, -1, fmt.Errorf("wtf?! Block not found for addr %X", addr)
}

func (b *Blockman) String() string {
	info := fmt.Sprintf("%d regions, start addr 0x%X, end addr 0x%X\n", len(b.blockRegions), b.baseAddr, b.endAddr)
	info += fmt.Sprintf("Total size: %d MB\n", (b.endAddr-b.baseAddr)/1024/1024)
	return info
}