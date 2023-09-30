package main

import (
	"fmt"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

func writeFlashFromFile(loader pmb887x.ChaosLoader, baseAddr, size int64, filePath string) error {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file to read flashdump: %v", err)
	}

	return loader.WriteFlash(baseAddr, buf[:size])
}
