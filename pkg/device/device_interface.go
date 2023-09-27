package device

type Device interface {
	// Name() returns a name and maybe some extra info about this Device. This info is not machine readable.
	Name() string
	// Connect() connects to the device. It may block. When it returns, a consumer can immediately read and write memory regions.
	Connect() error
	Disconnect() error
	SetSpeed(speed int) error
	ReadRegion(baseAddr, size int64) ([]byte, error)
	WriteRegion(baseAddr int64, block []byte) error
}
