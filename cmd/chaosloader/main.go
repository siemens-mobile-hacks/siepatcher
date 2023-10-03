package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

var (
	useEmulator   = flag.Bool("emulator", false, "Use emulator instead of a physical phone.")
	serialPort    = flag.String("serial", "", "Serial port path (like /dev/cu.usbserial-110, or COM2).")
	chaosLoader   = flag.String("loader", "", "Path to Chaos bootloader (.bin file).")
	chaosInfoFile = flag.String("chaos_info_file", "", "Path to a dumped Chaos info block. Parse and exit.")
	readFlash     = flag.Bool("read_flash", false, "Read flash to file.")
	writeFlash    = flag.Bool("write_flash", false, "Write flash from file.")
	flashFile     = flag.String("flash_file", "", "Path to a flash file to read from / store to.")
	flashBaseAddr = flag.Int64("base_addr", 0, "Base address to read from / write to.")
	flashLength   = flag.Int64("length", 0, "Length to read / to write.")
	applyPatch    = flag.Bool("apply_patch", false, "Apply patch specified by -patch_file.")
	revertPatch   = flag.Bool("revert_patch", false, "Revert patch specified by -patch_file.")
	dryRun        = flag.Bool("dry_run", false, "Only verify if a patch can be applied / reverted, but don't actually write data.")
	patchFile     = flag.String("patch_file", "", "Patch file to apply.")
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

	if *readFlash {
		if *flashBaseAddr == 0 || *flashLength == 0 || *flashFile == "" {
			fmt.Println("-base_addr, -length and -flash_file must be set!")
			os.Exit(1)
		}
		if err := readFlashToFile(chaos, *flashBaseAddr, *flashLength, *flashFile); err != nil {
			fmt.Printf("Cannot read flash from 0x%X len 0x%0X: %v", *flashBaseAddr, *flashLength, err)
			os.Exit(1)
		}
	}

	if *writeFlash {
		if *flashBaseAddr == 0 || *flashLength == 0 || *flashFile == "" {
			fmt.Println("-base_addr, -length and -flash_file must be set!")
			os.Exit(1)
		}
		if err := writeFlashFromFile(chaos, *flashBaseAddr, *flashLength, *flashFile); err != nil {
			fmt.Printf("Cannot write flash to 0x%X len 0x%0X: %v", *flashBaseAddr, *flashLength, err)
			os.Exit(1)
		}
	}

	if *applyPatch || *revertPatch {
		if err := DoApplyPatch(chaos, *patchFile, *revertPatch, *dryRun); err != nil {
			fmt.Printf("Cannot apply or revert patch %q! Error: %v", filepath.Base(*patchFile), err)
		}
	}
	dev.Disconnect()
	fmt.Println()
}
