package main

import (
	"fmt"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
)

func main() {
	fmt.Println("Emulator patcher")
	emu, err := device.NewEmulatorBackend()
	if err != nil {
		fmt.Printf("Error while creating an emulator device: %v", err)
		os.Exit(1)
	}
	fmt.Println("Waiting for connection from the emulator...")
	if err := emu.Connect(); err != nil {
		fmt.Printf("Error while connecting to emulator: %v", err)
		os.Exit(1)
	}

	fmt.Println("Emulator is ready and in Service Mode!")
}
