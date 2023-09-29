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
		f, err := os.Open(*chaosInfoFile)
		if err != nil {
			fmt.Printf("Cannot open info file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		info, err := pmb887x.ParseChaosInfo(f)
		if err != nil {
			fmt.Printf("Cannot parse information block: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(info)
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

	var info pmb887x.ChaosPhoneInfo
	if info, err = chaos.ReadInfo(); err != nil {
		fmt.Printf("Cannot read information from Chaos boot: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Phone information:\n%s\n", info)

	ff, err := os.Create("ff.bin")
	if err != nil {
		fmt.Printf("Cannot create fullflash file: %v", err)
		os.Exit(1)
	}
	defer ff.Close()
	bm := info.BlockMap
	maxRetries := 3
	stillNeedToRead := bm.TotalSize()
	readSize := 65536 // 64K
	errCount := 0
	baseAddr := bm.BaseAddr()
	totalRead := int64(0)
	for stillNeedToRead > 0 {
		buf := make([]byte, readSize)
		retries := maxRetries
		for ; retries > 0; retries-- {
			fmt.Printf("[Retry %d] Transfering %d bytes from addr %X...", maxRetries-retries, readSize, baseAddr)
			if err := chaos.ReadFlash(int64(baseAddr), buf); err != nil {
				fmt.Printf("\n\tError reading flash @ %08X: %v. Retries left: %d\n", baseAddr, err, retries)
				errCount++
				continue
			}
			break
		}
		if retries == 0 {
			fmt.Printf("\n\tCannot read block @ %08X after retries!\n", baseAddr)
			os.Exit(1)
		}

		n, err := ff.Write(buf)
		if n != len(buf) {
			fmt.Printf("\n\tCannot write block @ %08X to the fullflash file: %v\n", baseAddr, err)
			os.Exit(1)
		}
		fmt.Println("ok")
		baseAddr += int64(readSize)
		stillNeedToRead -= int64(readSize)
		totalRead += int64(readSize)
	}
	fmt.Printf("Ready! Transfered %d MB, error count: %d...\n", totalRead, errCount)
	dev.Disconnect()
	fmt.Println()
}
