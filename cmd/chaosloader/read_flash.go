package main

import (
	"fmt"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

func readFlashToFile(loader pmb887x.ChaosLoader, baseAddr, size int64, filePath string) error {
	ff, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file to store flashdump: %v", err)
	}
	defer ff.Close()

	maxRetries := 3
	stillNeedToRead := size
	readSize := 65536 // 64K
	errCount := 0

	totalRead := int64(0)
	for stillNeedToRead > 0 {
		buf := make([]byte, readSize)
		retries := maxRetries
		for ; retries > 0; retries-- {
			fmt.Printf("[Retry %d] Transfering %d bytes from addr %X...", maxRetries-retries, readSize, baseAddr)
			if err := loader.ReadFlash(int64(baseAddr), buf); err != nil {
				fmt.Printf("\n\tError reading flash @ %08X: %v. Retries left: %d\n", baseAddr, err, retries)
				errCount++
				continue
			}
			break
		}
		if retries == 0 {
			fmt.Printf("\n\tCannot read block @ %08X after retries!\n", baseAddr)
			return fmt.Errorf("cannot read block @ %08X after retries", baseAddr)
		}

		n, err := ff.Write(buf)
		if n != len(buf) {
			return fmt.Errorf("cannot write block @ %08X to the fullflash file: %v", baseAddr, err)
		}
		fmt.Println("ok")
		baseAddr += int64(readSize)
		stillNeedToRead -= int64(readSize)
		totalRead += int64(readSize)
	}
	return nil
}
