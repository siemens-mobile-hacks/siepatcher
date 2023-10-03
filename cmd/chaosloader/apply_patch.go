package main

import (
	"fmt"
	"log"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/patchreader"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

func DoApplyPatch(loader pmb887x.ChaosLoader, patchFile string, isRevert, isDryRun bool) error {
	// Load a patch.
	pr, err := patchreader.FromFile(patchFile)
	if err != nil {
		log.Fatalf("Cannot load patch: %v", err)
	}
	log.Printf("Loaded and parsed the patch successfully")

	flashInfo, err := loader.ReadInfo()
	if err != nil {
		return err
	}
	blockMapper := flashInfo.BlockMap
	var blockCache map[int64][]byte = map[int64][]byte{}
	// Figure out what blocks need to be modified.
	patchChunks := pr.Chunks()

	for _, chunk := range patchChunks {
		for addr := chunk.BaseAddr; addr < chunk.EndAddr(); addr++ {
			baseAddr, size, err := blockMapper.ParamsForAddr(addr + blockMapper.BaseAddr())
			if err != nil {
				log.Fatalf("Error when mapping patch chunks to blocks: %v", err)
			}
			if _, ok := blockCache[baseAddr]; ok {
				// This block is cached.
				continue
			}
			fmt.Printf("Need to request block @ %X size %X\n", baseAddr, size)
			blockCache[baseAddr] = make([]byte, size)
			if err := loader.ReadFlash(int64(baseAddr), blockCache[baseAddr]); err != nil {
				fmt.Printf("\n\tError reading flash @ %08X: %v\n", baseAddr, err)
			}
		}
	}

	for _, chunk := range patchChunks {
		for addr := chunk.BaseAddr; addr < chunk.EndAddr(); addr++ {
			// Get the base address of the block the current address is in.
			// This is also an index in our blockCache map.
			blockBaseAddr, _, _ := blockMapper.ParamsForAddr(addr + blockMapper.BaseAddr())
			// Offset inside the cached block.
			cachedBlockOff := blockMapper.BaseAddr() + addr - blockBaseAddr
			// Data offset inside the patch chunk.
			dataOff := addr - chunk.BaseAddr
			//fmt.Printf("\naddr=%08X, blockBaseAddr=%08X, cachedBlockOff=%08X, dataOff=%08X\n", addr, blockBaseAddr, cachedBlockOff, dataOff)
			gotOldData := &blockCache[blockBaseAddr][cachedBlockOff]
			var wantOldData, newData byte

			if !isRevert {
				wantOldData = chunk.OldData[dataOff]
				newData = chunk.NewData[dataOff]
			} else {
				wantOldData = chunk.NewData[dataOff]
				newData = chunk.OldData[dataOff]
			}
			if *gotOldData != wantOldData {
				log.Fatalf("Data at addr 0x%X is %X, expected %X", addr, *gotOldData, wantOldData)
			}
			*gotOldData = newData
		}
	}
	fmt.Println("Patch can be applied!")
	if isDryRun {
		return nil
	}

	// Now all blocks in our blockCache are patched.
	// Time to send them back to the phone.
	for addr, block := range blockCache {
		fmt.Printf("Writing block @ %08X len %08X\n", addr, len(block))
		if err := loader.WriteFlash(addr, block); err != nil {
			fmt.Printf("Error writing block: %v\n", err)
		}
	}

	fmt.Println("Patch applied!")

	return nil
}
