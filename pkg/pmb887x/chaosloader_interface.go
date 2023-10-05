package pmb887x

type ChaosLoaderInterface interface {
	Activate() error
	Ping() (bool, error)
	SetSpeed(speed int, speedSetter SpeedSetterFunc) error
	ReadInfo() (ChaosPhoneInfo, error)
	ReadFlash(baseAddr int64, buf []byte) error
	WriteFlash(baseAddr int64, buf []byte) error
}
