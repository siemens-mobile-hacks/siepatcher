package device

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

const (
	emuSock = "/tmp/siemens.sock"
)

type EmulatorDevice struct {
	socketPath string
	listener   net.Listener
	dev        pmb887x.Device
}

func NewEmulatorBackend() (*EmulatorDevice, error) {
	// Remove the socket file if it already exists
	if err := os.Remove(emuSock); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error removing existing socket: %v", err)
	}

	// Create a listener for the UNIX socket
	listener, err := net.Listen("unix", emuSock)
	if err != nil {
		return nil, fmt.Errorf("error creating socket listener %v", err)
	}

	return &EmulatorDevice{
		socketPath: emuSock,
		listener:   listener,
	}, nil
}

func (e *EmulatorDevice) Name() string {
	return fmt.Sprintf("pmb887x-emulator on %q", e.socketPath)
}

func (e *EmulatorDevice) Connect() error {

	log.Println("Waiting for emulator to connect")
	// This blocks until an emulator connects!
	conn, err := e.listener.Accept()
	if err != nil {
		return fmt.Errorf("cannot accept emulator connection: %v", err)
	}
	log.Printf("Emulator connected from %s", conn.RemoteAddr())

	e.dev = pmb887x.NewPMB(conn)

	return e.dev.LoadBoot()
}

func (e *EmulatorDevice) Disconnect() error {
	return e.dev.Disconnect()
}

func (e *EmulatorDevice) SetSpeed(speed int) error {
	return nil
}

func (e *EmulatorDevice) ReadRegion(baseAddr, size int64) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EmulatorDevice) WriteRegion(baseAddr int64, block []byte) error {
	return fmt.Errorf("not implemented")
}
