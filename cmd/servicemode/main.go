package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

var (
	useEmulator = flag.Bool("emulator", false, "Use emulator instead of a physical phone.")
	serialPort  = flag.String("serial", "", "Serial port path (like /dev/cu.usbserial-110, or COM2).")
)

func main() {

	var dev device.Device
	var err error

	flag.Parse()

	if *useEmulator {

		dev, err = device.NewEmulatorBackend()
		if err != nil {
			fmt.Printf("Cannot create new emulator connection: %v", err)
			os.Exit(1)
		}
	} else {
		if *serialPort == "" {
			fmt.Println("Must specify a serial port path")
			os.Exit(1)
		}
		dev, err = device.NewPhone(*serialPort)
		if err != nil {
			fmt.Printf("Cannot instantiate new phone connection: %v", err)
			os.Exit(1)
		}
	}

	if err = dev.ConnectAndBoot(pmb887x.ServiceModeBoot); err != nil {
		fmt.Printf("Cannot boot device into service mode: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s should be in Service Mode now!\n", dev.Name())
}
