package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

var (
	useEmulator   = flag.Bool("emulator", false, "Use emulator instead of a physical phone.")
	serialPort    = flag.String("serial", "", "Serial port path (like /dev/cu.usbserial-110, or COM2).")
	chaosLoader   = flag.String("loader", "", "Path to Chaos bootloader (.bin file).")
	chaosInfoFile = flag.String("chaos_info_file", "", "Path to a dumped Chaos info block. Parse and exit.")
)

func main() {

	var dev device.Device
	var err error

	flag.Parse()

	if *chaosInfoFile != "" {
		pmb887x.ParseChaosInfo(*chaosInfoFile)
		os.Exit(0)
	}

	if *useEmulator {

		dev, err = device.NewEmulatorBackend()
		if err != nil {
			fmt.Printf("Cannot create new emulator connection: %v\n", err)
			os.Exit(1)
		}
	} else {
		if *serialPort == "" {
			fmt.Println("Must specify a serial port path")
			os.Exit(1)
		}
		dev, err = device.NewPhone(*serialPort)
		if err != nil {
			fmt.Printf("Cannot instantiate new phone connection: %v\n", err)
			os.Exit(1)
		}
	}

	loader, err := os.ReadFile(*chaosLoader)
	if err != nil {
		fmt.Printf("cannot read Chaos Loader code: %v", err)
		os.Exit(1)
	}

	if err = dev.ConnectAndBoot(loader); err != nil {
		fmt.Printf("Cannot boot device with Chaos boot: %v", err)
		os.Exit(1)
	}

	// Now create a Chaos controller so  that all other operations interact with it
	// instead of a plain firmware.
	chaos := pmb887x.ChaosControllerForDevice(dev.PMB())

	if err = chaos.Activate(); err != nil {
		fmt.Printf("Cannot activate Chaos boot: %v", err)
		os.Exit(1)
	}

	if err = chaos.ReadInfo(); err != nil {
		fmt.Printf("Cannot read information from Chaos boot: %v", err)
		os.Exit(1)
	}

	dev.Disconnect()
}
