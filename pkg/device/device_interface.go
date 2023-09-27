package device

import "github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"

type Device interface {
	// Name() returns a name and maybe some extra info about this Device. This info is not machine readable.
	Name() string
	// Connect() connects to the device. It may block.
	ConnectAndBoot(loaderBin []byte) error
	Disconnect() error
	SetSpeed(speed int) error
	PMB() pmb887x.Device
}
