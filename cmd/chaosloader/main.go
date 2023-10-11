package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

var (
	useEmulator   = flag.Bool("emulator", false, "Use emulator instead of a physical phone.")
	useFullFlash  = flag.Bool("use_fullflash_not_phone", false, "Use a file with fullflash instead of a physical phone.")
	usedFFFile    = flag.String("use_fullflash_file_path", "", "Use this file instead of a real phone.")
	serialPort    = flag.String("serial", "", "Serial port path (like /dev/cu.usbserial-110, or COM2).")
	serialSpeed   = flag.Int("speed", 115200, "Serial port speed to use.")
	chaosLoader   = flag.String("loader", "", "Path to Chaos bootloader (.bin file).")
	useRestoreOld = flag.Bool("restore_old_data_from_ff", false, "If true, restore blocks changed by patch -patch_file from the FF backup -flash_file.")
	readFlash     = flag.Bool("read_flash", false, "Read flash to file.")
	writeFlash    = flag.Bool("write_flash", false, "Write flash from file.")
	flashFile     = flag.String("flash_file", "", "Path to a flash file to read from / store to.")
	flashBaseAddr = flag.Int64("base_addr", 0, "Base address to read from / write to.")
	flashLength   = flag.Int64("length", 0, "Length to read / to write.")
	applyPatch    = flag.Bool("apply_patch", false, "Apply patch specified by -patch_file.")
	revertPatch   = flag.Bool("revert_patch", false, "Revert patch specified by -patch_file.")
	dryRun        = flag.Bool("dry_run", false, "Only verify if a patch can be applied / reverted, but don't actually write data.")
	forceAction   = flag.Bool("force", false, "Apply /revert patch even if the old data doesn't match.")
	patchFile     = flag.String("patch_file", "", "Patch file to apply.")
)

func main() {

	var dev device.Device
	var chaos pmb887x.ChaosLoaderInterface
	var err error

	flag.Parse()

	if *useFullFlash {
		fullflash := device.NewDeviceFromFullflash(*usedFFFile)
		chaos = device.NewLoaderForFullflashFile(fullflash)
		dev = fullflash
	} else if *useEmulator {

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
		chaos = pmb887x.ChaosControllerForDevice(dev.PMB())
	}
	if err = chaos.Activate(); err != nil {
		fmt.Printf("Cannot activate Chaos boot: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Attempting to change COM speed to %d\n", *serialSpeed)
	ourSpeedSetter := func() error { return dev.SetSpeed(*serialSpeed) }
	if err := chaos.SetSpeed(*serialSpeed, ourSpeedSetter); err != nil {
		fmt.Printf("Cannot set comms speed %d with Chaos boot: %v\n", serialSpeed, err)
		os.Exit(1)
	}

	var info pmb887x.ChaosPhoneInfo
	if info, err = chaos.ReadInfo(); err != nil {
		fmt.Printf("Cannot read information from Chaos boot: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Phone information:\n%s\n", info)

	beginTime := time.Now()

	if *useRestoreOld {
		if *flashFile == "" || *patchFile == "" {
			log.Fatalf("-flash_file and -patch_file must not be empty!")
		}
		if err := RestoreOldDataFromFullflash(chaos, *patchFile, *flashFile); err != nil {
			log.Fatalf("Cannot restore data: %v", err)
		}
	}

	if *readFlash {
		if *flashBaseAddr == 0 || *flashLength == 0 || *flashFile == "" {
			fmt.Println("-base_addr, -length and -flash_file must be set!")
			os.Exit(1)
		}
		printScaryTimeStats()
		if err := readFlashToFile(chaos, *flashBaseAddr, *flashLength, *flashFile); err != nil {
			fmt.Printf("Cannot read flash from 0x%X len 0x%0X: %v\n", *flashBaseAddr, *flashLength, err)
			os.Exit(1)
		}
	}

	if *writeFlash {
		if *flashBaseAddr == 0 || *flashLength == 0 || *flashFile == "" {
			fmt.Println("-base_addr, -length and -flash_file must be set!")
			os.Exit(1)
		}
		printScaryTimeStats()
		if err := writeFlashFromFile(chaos, *flashBaseAddr, *flashLength, *flashFile); err != nil {
			fmt.Printf("Cannot write flash to 0x%X len 0x%0X: %v\n", *flashBaseAddr, *flashLength, err)
			os.Exit(1)
		}
	}

	if *applyPatch || *revertPatch {
		if err := DoApplyPatch(chaos, *patchFile, *revertPatch, *dryRun, *forceAction); err != nil {
			fmt.Printf("Cannot apply or revert patch %q! Error: %v", filepath.Base(*patchFile), err)
		}
	}
	elapsed := time.Since(beginTime)
	fmt.Printf("Operation took %v.\n", elapsed)
	dev.Disconnect()
	fmt.Println()
}

func printScaryTimeStats() {
	needTime := *flashLength / int64(*serialSpeed/8)
	dur, _ := time.ParseDuration(fmt.Sprintf("%ds", needTime))
	fmt.Printf("This operation will take %v of your life with the current serial port speed.\n", dur)
}
