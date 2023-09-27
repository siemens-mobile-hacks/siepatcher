package device

type Device interface {
	Name() string
	Connect() error
	Disconnect() error
	SetSpeed(speed int) error
	ReadRegion(baseAddr, size int64) ([]byte, error)
	WriteRegion(baseAddr int64, block []byte) error
}
