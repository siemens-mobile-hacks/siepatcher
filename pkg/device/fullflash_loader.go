package device

import (
	"github.com/siemens-mobile-hacks/siepatcher/pkg/blockman"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

// FullflashLoader implements ChaosLoaderInterface.
type FullflashLoader struct {
	ff *FullflashFile
	bm blockman.Blockman
}

func NewLoaderForFullflashFile(ff *FullflashFile) *FullflashLoader {
	return &FullflashLoader{ff: ff}
}

func (fl *FullflashLoader) Activate() error {
	return fl.ff.ConnectAndBoot(nil)
}

func (fl *FullflashLoader) Ping() (bool, error) {
	return true, nil
}

func (fl *FullflashLoader) SetSpeed(speed int, speedSetter pmb887x.SpeedSetterFunc) error {
	return nil
}

func (fl *FullflashLoader) ReadInfo() (pmb887x.ChaosPhoneInfo, error) {
	// Create a blockmap because a higher-level code needs it.
	fl.bm = blockman.New(0xA0000000)
	blockSize := int64(0x20000)
	blockCount := int(fl.ff.Size() / blockSize)
	fl.bm.AddRegion(blockSize, blockCount)

	return pmb887x.ChaosPhoneInfo{
		ModelName:    "Fullflash dump",
		Manufacturer: "siemens-mobile-hacks Org",
		IMEI:         "xxxxxxxxxxxxxxx",
		BlockMap:     fl.bm,
	}, nil
}

func (fl *FullflashLoader) ReadFlash(baseAddr int64, buf []byte) error {
	sizeToRead := len(buf)
	readBuf, err := fl.ff.ReadRegion(baseAddr-fl.bm.BaseAddr(), int64(sizeToRead))
	if err != nil {
		return err
	}
	copy(buf, readBuf)
	return nil
}

func (fl *FullflashLoader) WriteFlash(baseAddr int64, buf []byte) error {
	return fl.ff.WriteRegion(baseAddr-fl.bm.BaseAddr(), buf)
}
