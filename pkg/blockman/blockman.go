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
		endAddr:      baseAddr - 1,
		totalSize:    0,
		blockRegions: []BlockRegionInfo{},
	}
}

func (b *Blockman) TotalSize() int64 {
	return b.totalSize
}

func (b *Blockman) BaseAddr() int64 {
	return b.baseAddr
}

func (b *Blockman) EndAddr() int64 {
	return b.endAddr
}

func (b *Blockman) NumOfRegions() int {
	return len(b.blockRegions)
}

func (b *Blockman) AddRegion(blockSize int64, blockCount int) error {
	var regionSize int64 = blockSize * int64(blockCount)
	blockRegion := BlockRegionInfo{
		baseAddr:   b.endAddr + 1,
		endAddr:    b.endAddr + 1 + regionSize - 1,
		blockSize:  blockSize,
		blockCount: blockCount,
	}
	b.blockRegions = append(b.blockRegions, blockRegion)
	b.endAddr += int64(regionSize)
	b.totalSize += int64(regionSize)
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

// String implements Stringer interface.
func (b Blockman) String() string {
	info := fmt.Sprintf("Total size: %d MB\n", b.totalSize/1024/1024)
	info += fmt.Sprintf("%d regions, start addr 0x%X, end addr 0x%X; total size 0x%08X\n",
		len(b.blockRegions), b.baseAddr, b.endAddr, b.totalSize)
	for i, reg := range b.blockRegions {
		info += fmt.Sprintf("  Region #%d: [%08X, %08X] %d blocks, size of each 0x%X\n", i, reg.baseAddr, reg.endAddr, reg.blockCount, reg.blockSize)
	}
	return info
}
