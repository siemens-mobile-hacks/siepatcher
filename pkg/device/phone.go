package device

import (
	"fmt"
	"time"

	"github.com/kibab/goserial"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

// Phone represents a real device connected to a serial port.
type Phone struct {
	serialPath string
	serialPort *goserial.Port
	dev        pmb887x.Device
}

func NewPhone(serialPortNameOrPath string) (*Phone, error) {
	serialPortConfig := &goserial.Config{Name: serialPortNameOrPath, Baud: 115200, ReadTimeout: time.Second * 5}
	serialPort, err := goserial.OpenPort(serialPortConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot open serial port %q: %v", serialPortNameOrPath, err)
	}

	pmb := pmb887x.NewPMB(serialPort)
	return &Phone{
		dev:        pmb,
		serialPort: serialPort,
		serialPath: serialPortNameOrPath,
	}, nil
}

func (p *Phone) Name() string {
	return fmt.Sprintf("Real phone at %q", p.serialPath)
}

func (p *Phone) ConnectAndBoot(loaderBin []byte) error {
	if err := p.dev.LoadBoot(loaderBin); err != nil {
		return err
	}
	return nil
}

func (p *Phone) Disconnect() error {
	return p.dev.Disconnect()
}

func (p *Phone) SetSpeed(speed int) error {
	return p.serialPort.SetSpeed(speed)
}

func (p *Phone) PMB() pmb887x.Device {
	return p.dev
}
