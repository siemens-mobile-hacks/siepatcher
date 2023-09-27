package pmb887x

import (
	"fmt"
	"io"
	"log"
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

// LoadBoot initializes PMB serial communication and sends the bootloader.
func (pmb *Device) LoadBoot() error {
	log.Println("Initializing connection")
	if _, err := pmb.iostream.Write([]byte("ATAT")); err != nil {
		return fmt.Errorf("error writing to client: %v", err)
	}

	var buf []byte = make([]byte, 1)
	n, err := pmb.iostream.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading from client: %v", err)
	}
	log.Printf("Read %d bytes", n)
	phoneType := buf[0]

	var phoneTypeStr string
	switch phoneType {
	case 0xB0:
		phoneTypeStr = "SGOLD"
	case 0xC0:
		phoneTypeStr = "SGOLD2"
	default:
		return fmt.Errorf("unknown phone type %x", phoneType)
	}
	log.Printf("Phone type: %s", phoneTypeStr)

	// Prepare payload.
	ldrLen := len(serviceModeBoot)
	payload := []byte{0x30, byte(ldrLen & 0xFF), byte((ldrLen >> 8) & 0xFF)}
	var chk byte = 0
	for i := 0; i < ldrLen; i++ {
		var b byte = serviceModeBoot[i]
		chk ^= b
		payload = append(payload, b)
	}
	payload = append(payload, chk)

	log.Printf("Generated loader payload len %d", len(payload))
	//fmt.Printf("%s\n", hex.Dump(payload))

	// Send payload.
	log.Println("Sending payload")
	for i := 0; i < len(payload); i++ {
		if _, err := pmb.iostream.Write([]byte{payload[i]}); err != nil {
			return fmt.Errorf("error writing payload: %v", err)
		}
		fmt.Print(".")
	}
	fmt.Println()

	fmt.Println("Waiting for ACK")
	n, err = pmb.iostream.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading from client: %v", err)
	}
	log.Printf("Read %d bytes", n)
	ack := buf[0]

	if !(ack == 0xC1 || ack == 0xB1) {
		return fmt.Errorf("uknown ack byte %x", ack)
	}
	log.Println("Boot code loaded!")
	return nil
}

func (pmb *Device) Disconnect() error {
	return pmb.iostream.Close()
}
