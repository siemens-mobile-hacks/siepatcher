package main

import (
	"fmt"
	"log"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

var dev device.Device
var chaos pmb887x.ChaosLoaderInterface
var err error

func errReply(errstr error, reply chan<- PatcherReply) {
	log.Print(errstr)
	rep := PatcherReply{
		EventType:  CmdError,
		ErrorDescr: errstr.Error(),
	}
	reply <- rep
}

func reportProgress(msg string, reply chan<- PatcherReply) {
	log.Print(msg)
	rep := PatcherReply{
		EventType:     CmdProgress,
		ProgressDescr: msg,
	}
	reply <- rep
}

func PatchEngine(cmd <-chan PatcherCommand, reply chan<- PatcherReply) {
	for ev := range cmd {
		log.Printf("PatchEngine: Got a command %v", ev)
		switch ev.EventType {
		case ConnectTarget:
			if ev.ConnectInfo.SerialPath != "" {
				dev, err = device.NewPhone(ev.ConnectInfo.SerialPath)
				if err != nil {
					errReply(fmt.Errorf("cannot instantiate new phone connection: %v", err), reply)
					dev.Disconnect()
					continue
				}

				reportProgress("Press RED button", reply)

				if err = dev.ConnectAndBoot(pmb887x.ChaosLoaderBin); err != nil {
					errReply(fmt.Errorf("cannot boot device with Chaos boot: %v", err), reply)
					dev.Disconnect()
					continue
				}

				// Now create a Chaos controller so  that all other operations interact with it
				// instead of a plain firmware.
				chaos = pmb887x.ChaosControllerForDevice(dev.PMB())
			} else if ev.ConnectInfo.EmuSocketPath != "" {
				errReply(fmt.Errorf("not implemented"), reply)
				continue
			} else if ev.ConnectInfo.FFPath != "" {
				errReply(fmt.Errorf("not implemented"), reply)
				continue
			}

			if err = chaos.Activate(); err != nil {
				errReply(fmt.Errorf("cannot activate Chaos boot"), reply)
				dev.Disconnect()
				continue
			}

			// fmt.Printf("Attempting to change COM speed to %d\n", *serialSpeed)
			// ourSpeedSetter := func() error { return dev.SetSpeed(*serialSpeed) }
			// if err := chaos.SetSpeed(*serialSpeed, ourSpeedSetter); err != nil {
			// 	fmt.Printf("Cannot set comms speed %d with Chaos boot: %v\n", serialSpeed, err)
			// 	os.Exit(1)
			// }

			var info pmb887x.ChaosPhoneInfo
			if info, err = chaos.ReadInfo(); err != nil {
				errReply(fmt.Errorf("cannot read information from Chaos boot: %v", err), reply)
				dev.Disconnect()
				continue
			}

			rep := PatcherReply{
				EventType: TargetInfo,
				DeviceInfo: struct{ PhoneInfo pmb887x.ChaosPhoneInfo }{
					PhoneInfo: info,
				},
			}
			reply <- rep
		}
	}
}
