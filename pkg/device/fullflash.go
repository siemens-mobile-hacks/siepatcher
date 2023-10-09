package device

import (
	"fmt"
	"io"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

type writeableBackingStore interface {
	io.ReadWriteCloser
	io.ReadWriteSeeker
}

// FullflashFile represents a file with the phone flash dump (fullflash) on disk.
// It implements Device interface.
type FullflashFile struct {
	backingStore writeableBackingStore
	fileName     string
	fileSize     int64
}

// NewDeviceFromFullflash creates an instance of FullflashFile.
func NewDeviceFromFullflash(filePath string) *FullflashFile {
	return &FullflashFile{
		fileName: filePath,
	}
}

func (ff *FullflashFile) Name() string {
	return fmt.Sprintf("Flash dump file %q", ff.fileName)
}

func (ff *FullflashFile) PMB() pmb887x.Device {
	return pmb887x.Device{}
}

func (ff *FullflashFile) ConnectAndBoot(_ []byte) error {
	if ff.backingStore != nil {
		return fmt.Errorf("file %q is already open", ff.fileName)
	}

	f, err := os.OpenFile(ff.fileName, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	ff.backingStore = f

	// Now obtain the file size so that we initialize a block map.
	fileStat, err := f.Stat()
	if err != nil {
		return err
	}
	ff.fileSize = fileStat.Size()
	return nil
}

func (ff *FullflashFile) Disconnect() error {
	if ff.backingStore == nil {
		return fmt.Errorf("already disconnected")
	}
	if err := ff.backingStore.Close(); err != nil {
		return err
	}
	ff.backingStore = nil
	ff.fileSize = 0
	return nil
}

func (ff *FullflashFile) SetSpeed(speed int) error {
	return nil
}

func (ff *FullflashFile) ReadRegion(baseAddr, size int64) ([]byte, error) {
	if ff.backingStore == nil {
		return nil, fmt.Errorf("need to connect first")
	}

	if baseAddr+size > ff.fileSize {
		return nil, fmt.Errorf("region [0x%x, 0x%x] out of bounds (max addr %x)", baseAddr, baseAddr+size-1, ff.fileSize)
	}

	if _, err := ff.backingStore.Seek(baseAddr, 0); err != nil {
		return nil, fmt.Errorf("cannot seek to 0x%x: %v", baseAddr, err)
	}

	buf := make([]byte, size)
	n, err := ff.backingStore.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read %d bytes: %v", size, err)
	}
	if n < int(size) {
		return nil, fmt.Errorf("got %d bytes, want %d", n, size)
	}
	return buf, nil
}

func (ff *FullflashFile) WriteRegion(baseAddr int64, block []byte) error {
	if ff.backingStore == nil {
		return fmt.Errorf("need to connect first")
	}

	bytesToWrite := int64(len(block))
	if baseAddr+bytesToWrite > ff.fileSize {
		return fmt.Errorf("region [0x%x, 0x%x] out of bounds (max addr %x)", baseAddr, baseAddr+bytesToWrite-1, ff.fileSize)
	}

	if _, err := ff.backingStore.Seek(baseAddr, 0); err != nil {
		return fmt.Errorf("cannot seek to 0x%x: %v", baseAddr, err)
	}

	n, err := ff.backingStore.Write(block)
	if err != nil {
		return fmt.Errorf("cannot write %d bytes: %v", len(block), err)
	}
	if int64(n) < bytesToWrite {
		return fmt.Errorf("got %d bytes, want %d", n, bytesToWrite)
	}

	return nil
}

// //////////////////////////////////
// FullflashFile specific methods //
// //////////////////////////////////
func (ff *FullflashFile) Size() int64 {
	return ff.fileSize
}
