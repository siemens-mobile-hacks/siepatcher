package device

import (
	"os"
	"testing"
)

func newDeviceFromBlankFile(size int64) (*FullflashFile, func()) {
	ff, err := os.CreateTemp("", "fullflash_*")
	if err != nil {
		panic("cannot create temp file?!")
	}
	buf := make([]byte, size)
	for i := int64(0); i < size; i++ {
		buf[i] = 0xFF
	}
	_, err = ff.Write(buf)
	if err != nil {
		panic("cannot write temp file?!")
	}

	return NewDeviceFromFullflash(ff.Name()), func() {
		if err := os.Remove(ff.Name()); err != nil {
			panic(err)
		}
	}
}

func TestFullflashFile(t *testing.T) {

	testFF, cleanup := newDeviceFromBlankFile(16)
	defer cleanup()

	if err := testFF.ConnectAndBoot(nil); err != nil {
		t.Fatalf("Error while initializing a test fullflash: %v", err)
	}

	buf0 := []byte{'W', 'T', 'F'}
	if err := testFF.WriteRegion(0, buf0); err != nil {
		t.Fatalf("Cannot WriteRegion(): %v", err)
	}

	if err := testFF.WriteRegion(14, buf0); err == nil {
		t.Fatalf("Expected WriteRegion() to fail")
	}

	rbuf, err := testFF.ReadRegion(2, 2)
	if err != nil {
		t.Fatalf("Cannot ReadRegion(): %v", err)
	}

	if !(rbuf[0] == 'F' && rbuf[1] == 0xFF) {
		t.Fatalf("Unexpected read buffer contents: %v", rbuf)
	}

	_, err = testFF.ReadRegion(15, 2)
	if err == nil {
		t.Fatalf("Expected ReadRegion() to fail")
	}

}
