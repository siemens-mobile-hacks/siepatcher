package pmb887x

import (
	"fmt"
	"io"
	"log"
	"time"
)

// Device is an entity that can run bootloaders and interact with us via
// a simple bi-directional stream of bytes.
type Device struct {
	iostream io.ReadWriteCloser
}

func NewPMB(io io.ReadWriteCloser) Device {
	return Device{
		iostream: io,
	}
}

func shortDelay() {
	time.Sleep(100 * time.Millisecond)
}

// LoadBoot initializes PMB serial communication and sends the bootloader.
func (pmb *Device) LoadBoot(bootcode []byte) error {
	log.Println("Initializing connection")

	var buf []byte = make([]byte, 1)
	var deviceType byte
	fmt.Println("Press RED button!")
	stopAT := false

	// Start spamming our device with a bunch of ATs.
	go func() {
		for {
			if stopAT {
				return
			}
			if _, err := pmb.iostream.Write([]byte("AT")); err != nil {
				fmt.Printf("error writing to client: %v", err)
			}
			fmt.Print(".")
			shortDelay()
		}
	}()

	// Read a phone type from the interface.
	for {
		_, err := pmb.iostream.Read(buf)
		if err != nil {
			return fmt.Errorf("error reading from client: %v", err)
		}
		deviceType = buf[0]
		if deviceType == 0xB0 || deviceType == 0xC0 {
			fmt.Println("\nConnected!")
			stopAT = true
			break
		}
	}
	var deviceTypeStr string
	switch deviceType {
	case 0xB0:
		deviceTypeStr = "SGOLD"
	case 0xC0:
		deviceTypeStr = "SGOLD2"
	default:
		return fmt.Errorf("unknown device type %x", deviceType)
	}
	log.Printf("Device type: %s", deviceTypeStr)

	// Prepare payload.
	ldrLen := len(bootcode)
	payload := []byte{0x30, byte(ldrLen & 0xFF), byte((ldrLen >> 8) & 0xFF)}
	var chk byte = 0
	for i := 0; i < ldrLen; i++ {
		var b byte = bootcode[i]
		chk ^= b
		payload = append(payload, b)
	}
	payload = append(payload, chk)

	log.Printf("Generated loader payload len %d", len(payload))

	// Send payload.
	log.Println("Sending payload")
	for i := 0; i < len(payload); i++ {
		if _, err := pmb.iostream.Write([]byte{payload[i]}); err != nil {
			return fmt.Errorf("error writing payload: %v", err)
		}
		fmt.Print(".")
	}
	fmt.Println()

	// Give bootloader some time to init.
	shortDelay()

	fmt.Println("Waiting for ACK")
	n, err := pmb.iostream.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading from client: %v", err)
	}
	log.Printf("Read %d bytes", n)
	ack := buf[0]

	if ack == 0x1C || ack == 0x1B {
		return fmt.Errorf("bootcode rejected by firmware (%x)", ack)
	}
	if !(ack == 0xC1 || ack == 0xB1) {
		return fmt.Errorf("uknown ack byte %x", ack)
	}
	log.Println("Boot code loaded")
	return nil
}

func (pmb *Device) Disconnect() error {
	return pmb.iostream.Close()
}
