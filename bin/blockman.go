package main

import (
	"fmt"
	"log"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/blockman"
)

func main() {
	log.Println("Blockman started")

	bm := blockman.New(0xA0000000)
	bm.AddRegion(0x1000000, 8)
	bm.AddRegion(0x0800000, 16)

	fmt.Println("Blocks info:")
	fmt.Println(bm.String())
}
