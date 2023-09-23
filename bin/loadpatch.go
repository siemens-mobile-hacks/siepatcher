package main

import (
	"log"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/patchreader"
)

func main() {
	log.Println("LoadPatch started")
	if len(os.Args) < 2 {
		log.Fatal("No file specified")
	}
	pr, err := patchreader.FromFile(os.Args[1])
	if err != nil {
		log.Fatalf("Cannot load patch: %v", err)
	}
	log.Printf("Loaded patch %s", pr)
}
