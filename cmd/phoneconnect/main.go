package main

import (
	"fmt"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/device"
)

func main() {
	serialPortPath := "/dev/cu.usbserial-110"

	phone, err := device.NewPhone(serialPortPath)

	if err != nil {
		fmt.Printf("Cannot initialize phone connection: %v\n", err)
		return
	}

	fmt.Printf("Phone initialized: %s\n", phone.Name())
	if err := phone.Connect(); err != nil {
		fmt.Printf("Cannot connect to phone: %v\n", err)
		return
	}

	phone.Disconnect()
}
