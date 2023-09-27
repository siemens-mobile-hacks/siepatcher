package pmb887x

import (
	"encoding/hex"
	"fmt"
	"log"
)

type ChaosLoader struct {
	pmb Device
}

func ChaosControllerForDevice(dev Device) ChaosLoader {
	return ChaosLoader{pmb: dev}
}

func (cl *ChaosLoader) Activate() error {
	// We need to get an ACK that Chaos boot loaded: 0xA5
	r := []byte{0x0}
	if _, err := cl.pmb.iostream.Read(r); err != nil {
		return fmt.Errorf("error reading chaos loader ready message: %v", err)
	}
	if r[0] != 0xA5 {
		return fmt.Errorf("unknown chaos loader ready message %X", r[0])
	}
	shortDelay()

	// We need to send one ping to activate loader.
	pong, err := cl.Ping()
	if err != nil {
		return fmt.Errorf("error sending first ping: %v", err)
	}
	if !pong {
		return fmt.Errorf("chaos didn't reply to the first ping")
	}
	log.Print("Chaos bootloader activated")
	return nil
}

func (cl *ChaosLoader) Ping() (bool, error) {
	if _, err := cl.pmb.iostream.Write([]byte{'A'}); err != nil {
		return false, err
	}
	shortDelay()
	reply := []byte{0x00}
	if _, err := cl.pmb.iostream.Read(reply); err != nil {
		return false, err
	}
	if reply[0] == 'R' {
		return true, nil
	}
	return false, nil
}

func (cl *ChaosLoader) ReadInfo() error {
	fmt.Println("Requesting information")
	shortDelay()
	if _, err := cl.pmb.iostream.Write([]byte{'I'}); err != nil {
		return err
	}
	shortDelay()
	reply := make([]byte, 128)
	var n int
	var err error
	if n, err = cl.pmb.iostream.Read(reply); err != nil {
		return err
	}
	fmt.Printf("len=%d bytes:\n%s\n", n, hex.Dump(reply[0:n]))
	if n < 128 {
		return fmt.Errorf("less than 128 bytes read: %d", n)
	}
	return nil
}
