package main

import (
	"fmt"
	"log"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/blockman"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/patchreader"
)

func main() {
	log.Println("LoadPatch started")
	if len(os.Args) < 2 {
		log.Fatal("No file specified")
	}

	// Load a patch.
	pr, err := patchreader.FromFile(os.Args[1])
	if err != nil {
		log.Fatalf("Cannot load patch: %v", err)
	}
	log.Printf("Loaded and parsed the patch successfully")

	// Initialize a block map.
	var flashStartAddr int64 = 0xA0000000
	blockMapper := blockman.New(flashStartAddr)

	/*
		C81:
		#0: 255 blocks x 131072 bytes
		#1: 4 blocks x 32768 bytes

		#0: 4 blocks x 32768 bytes
		#1: 255 blocks x 131072 bytes
	*/
	/*
		blockMapper.AddRegion(131072, 255)
		blockMapper.AddRegion(32768, 4)

		blockMapper.AddRegion(32768, 4)
		blockMapper.AddRegion(131072, 255)
	*/

	// A random phone, flash info from our chat.
	// Bank 0
	blockMapper.AddRegion(131072, 255)
	blockMapper.AddRegion(32768, 4)
	// Bank 1
	blockMapper.AddRegion(131072, 255)
	blockMapper.AddRegion(32768, 4)

	fmt.Println("Flash Blocks info:")
	fmt.Println(blockMapper.String())

	var blockCache map[int64]int64 = map[int64]int64{}
	// Figure out what blocks need to be modified.
	patchChunks := pr.Chunks()

	for _, chunk := range patchChunks {
		for addr := chunk.BaseAddr; addr < chunk.EndAddr(); addr++ {
			baseAddr, size, err := blockMapper.ParamsForAddr(addr + flashStartAddr)
			if err != nil {
				log.Fatalf("Error when mapping patch chunks to blocks: %v", err)
			}
			if _, ok := blockCache[baseAddr]; ok {
				// This block is cached.
				continue
			}
			blockCache[baseAddr] = size
			log.Printf("Need to request block @ %X size %X", baseAddr, size)
		}
	}

	fullFlash, err := os.ReadFile("/tmp/fullflash.bin")
	if err != nil {
		log.Fatalf("Cannot load FF: %v", err)
	}

	for _, chunk := range patchChunks {
		for addr := chunk.BaseAddr; addr < chunk.EndAddr(); addr++ {
			dataOff := addr - chunk.BaseAddr
			if fullFlash[addr] != chunk.OldData[dataOff] {
				log.Fatalf("Data at addr 0x%X is %X, expected %X", addr, fullFlash[addr], chunk.OldData[dataOff])
			}
		}
	}
	fmt.Println("Patch can be applied!")
}
