package main

import (
	"fmt"
	"net"
	"os"
)

var (
	serviceModeBoot []byte = []byte{0xF1, 0x04, 0xA0, 0xE3, 0x20, 0x10, 0x90, 0xE5, 0xFF, 0x10, 0xC1, 0xE3, 0xA5, 0x10, 0x81, 0xE3,
		0x20, 0x10, 0x80, 0xE5, 0x1E, 0xFF, 0x2F, 0xE1, 0x04, 0x01, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x53, 0x49, 0x45, 0x4D, 0x45, 0x4E, 0x53, 0x5F, 0x42, 0x4F, 0x4F, 0x54, 0x43, 0x4F, 0x44, 0x45,
		0x01, 0x00, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x04, 0x05, 0x00, 0x8B, 0x00, 0x8B}
)

func handleConnection(conn net.Conn) {

	defer conn.Close()

	// Handle incoming data or requests here
	_, err := conn.Write([]byte("ATAT"))
	if err != nil {
		fmt.Println("Error writing to client:", err)
		return
	}

	var buf []byte = make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}
	fmt.Printf("Read %d bytes\n", n)
	phoneType := buf[0]

	fmt.Printf("Phone type: %X\n", phoneType)
	switch phoneType {
	case 0xB0:
		fmt.Println("SGOLD")
	case 0xC0:
		fmt.Println("SGOLD2")
	default:
		fmt.Println("Unknown!")
		return
	}

	// Prepare payload.
	ldrLen := len(serviceModeBoot)
	payload := []byte{0x30, byte(ldrLen & 0xFF), byte((ldrLen >> 8) & 0xFF)}
	var chk byte = 0
	for i := 0; i < ldrLen; i++ {
		var b byte = serviceModeBoot[i]
		chk ^= b
		payload = append(payload, b)
	}
	payload = append(payload, chk)

	fmt.Printf("Generated payload len %d\n", len(payload))
	//fmt.Printf("%s\n", hex.Dump(payload))

	// Send payload.
	fmt.Println("Sending payload:")
	for i := 0; i < len(payload); i++ {
		_, err := conn.Write([]byte{payload[i]})
		if err != nil {
			fmt.Println("Error writing payload:", err)
			return
		}
		fmt.Print(".")
	}
	fmt.Println()

	fmt.Println("Waiting for ACK")
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}
	fmt.Printf("Read %d bytes\n", n)
	ack := buf[0]

	if !(ack == 0xC1 || ack == 0xB1) {
		fmt.Printf("Uknown ack byte %x", ack)
		return
	}
	fmt.Println("Boot code loaded!")
}

func main() {
	// Define the path to the UNIX socket
	socketPath := "/tmp/siemens.sock"

	// Remove the socket file if it already exists
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error removing existing socket:", err)
		return
	}

	// Create a listener for the UNIX socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error creating socket listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on UNIX socket:", socketPath)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection in a goroutine
		go handleConnection(conn)
	}
}
