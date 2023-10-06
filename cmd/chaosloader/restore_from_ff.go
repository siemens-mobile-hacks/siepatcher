package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/patcheskibabcom"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/patchreader"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

func RestoreOldDataFromFullflash(loader pmb887x.ChaosLoaderInterface, patchFile, fullflashPath string) error {
	var pr *patchreader.PatchReader
	// Load a patch.
	patchID, err := strconv.ParseInt(patchFile, 10, 64)
	if err != nil {
		pr, err = patchreader.FromFile(patchFile)
	} else {
		patchText, err := patcheskibabcom.PatchByID(int(patchID))
		if err != nil {
			return err
		}
		if pr, err = patchreader.FromString(patchText); err != nil {
			return err
		}
	}
	if err != nil {
		return fmt.Errorf("cannot load patch: %v", err)
	}
	log.Printf("Loaded and parsed the patch successfully")

	ff := device.NewDeviceFromFullflash(fullflashPath)
	if err := ff.ConnectAndBoot(nil); err != nil {
		return fmt.Errorf("cannot load fullflash: %v", err)
	}

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
				return fmt.Errorf("error when mapping patch chunks to blocks: %v", err)
			}
			if _, ok := blockCache[baseAddr]; ok {
				// This block is cached.
				continue
			}
			fmt.Printf("Need to request block @ %X size %X\n", baseAddr, size)
			blockCache[baseAddr] = make([]byte, size)

			blockCache[baseAddr], err = ff.ReadRegion(int64(baseAddr-blockMapper.BaseAddr()), size)
			if err != nil {
				return fmt.Errorf("cannot read block @ %08X size %08X from fullflash: %v", baseAddr, size, err)
			}
		}
	}

	for addr, block := range blockCache {
		fmt.Printf("Writing block @ %08X len %08X\n", addr, len(block))
		if err := loader.WriteFlash(addr, block); err != nil {
			return fmt.Errorf("error writing block: %v", err)
		}
	}

	return nil
}
