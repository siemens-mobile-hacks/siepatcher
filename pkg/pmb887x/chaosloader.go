package pmb887x

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/siemens-mobile-hacks/siepatcher/pkg/blockman"
)

/*
Sources:
https://github.com/Azq2/pmb887x-dev/blob/master/chaos-boot.pl#L476
https://github.com/siemens-mobile-hacks/v-klay/blob/master/VDevicePhone.h#L84

#	BYTE strModelName[16];					// - model
#	BYTE strManufacturerName[16];			// - manufacturer
#	BYTE strIMEI[16];						//- IMEI (in ASCII)
#	BYTE reserved0[16];						// - (reserved)
#	DWORD flashBaseAddr;					// - base address of flash (ROM)
#	BYTE reserved1[12];						// - (reserved)
#	DWORD flash0Type;						//flash1 IC Manufacturer (LOWORD) and device ID (HIWORD)
#	BYTE flashSizePow;						// - N, CFI byte 27h. Size of flash = 2^N
#	WORD writeBufferSize;					// - CFI bytes 2Ah-2Bh size of write-buffer (not used by program)
#	BYTE flashRegionsNum;					// - CFI byte 2Ch - number of regions.
#	WORD flashRegion0BlocksNumMinus1;		// - N, CFI number of blocks in 1st region = N+1
#	WORD flashRegion0BlockSizeDiv256;		// - N, CFI size of blocks in 1st region = N*256
#	WORD flashRegion1BlocksNumMinus1;		// - N, CFI number of blocks in 2nd region = N+1
#	WORD flashRegion1BlockSizeDiv256;		// - N, CFI size of blocks in 2nd region = N*256
#	BYTE reserved2[32];						// - (reserved)

My EL71_2:
00000000  45 4c 37 31 00 00 00 00  00 00 00 00 00 00 00 00  |EL71............|
00000010  53 49 45 4d 45 4e 53 00  00 00 00 00 00 00 00 00  |SIEMENS.........|
00000020  XX XX XX XX XX XX XX XX  XX XX XX XX XX XX XX 00  |XXXXXXXXXXXXXXX.|
00000030  10 7a 5d 80 b0 c0 b4 5a  c3 48 d6 45 73 00 ae 0e  |.z]....Z.H.Es...|
00000040  00 00 00 a0 95 16 95 75  00 03 00 00 00 00 00 00  |.......u........|
00000050  89 00 7e 88 01 0a 02 01  ff 00 00 04 ff ff ff ff  |..~.............|
00000060  ff ff ff ff ff ff ff ff  00 00 00 00 00 00 00 00  |................|
00000070  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|

From Azq2:
00000000  45 4c 37 31 00 00 00 00  00 00 00 00 00 00 00 00  |EL71............|
00000010  53 49 45 4d 45 4e 53 00  00 00 00 00 00 00 00 00  |SIEMENS.........|
00000020  XX XX XX XX XX XX XX XX  XX XX XX XX XX XX XX 00  |XXXXXXXXXXXXXXX.|
00000030  8f 77 47 3e 07 43 3b 6a  6a a7 a8 bc 42 17 bd 5a  |.wG>.C;jj...B..Z|
00000040  00 00 00 a0 a9 75 dc 16  00 03 00 00 00 00 00 00  |.....u..........|
00000050  20 00 19 88 01 0a 02 01  ff 00 00 04 ff ff ff ff  | ...............|
00000060  ff ff ff ff ff ff ff ff  00 00 00 00 00 00 00 00  |................|
00000070  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
*/

// ChaosPhoneInfo holds information that we got and parsed from Chaos bootloader running on the device.
type ChaosPhoneInfo struct {
	ModelName    string
	Manufacturer string
	IMEI         string
	BlockMap     blockman.Blockman
}

// String implements fmt.Stringer.
func (i ChaosPhoneInfo) String() string {
	return fmt.Sprintf("Model %s by %s, IMEI %s\nFlash map:\n%s", i.ModelName, i.Manufacturer, i.IMEI, i.BlockMap)
}

// chaosInfo describes the on-the-wire format of reply to "info" command.
type chaosInfo struct {
	ModelName                   [16]byte
	Manufacturer                [16]byte
	IMEI                        [16]byte
	Reserved0                   [16]byte
	FlashBaseAddr               uint32
	Reserved1                   [12]byte
	Flash0Type                  uint32
	FlashSizePow                byte
	WriteBufSize                uint16
	FlashRegionsNum             byte
	FlashRegion0BlocksNumMinus1 uint16
	FlashRegion0BlockSizeDiv256 uint16
	FlashRegion1BlocksNumMinus1 uint16
	FlashRegion1BlockSizeDiv256 uint16
	FlashRegion2BlocksNumMinus1 uint16
	FlashRegion2BlockSizeDiv256 uint16
	FlashRegion3BlocksNumMinus1 uint16
	FlashRegion3BlockSizeDiv256 uint16
}

// readCmd is a on-wire format of Read Flash command.
type readCmd struct {
	Cmd  byte
	Addr uint32
	Size uint32
}

type ChaosLoader struct {
	pmb Device
	bm  *blockman.Blockman
}

func ChaosControllerForDevice(dev Device) ChaosLoader {
	return ChaosLoader{pmb: dev}
}

func (cl *ChaosLoader) Activate() error {
	// We need to get an ACK that Chaos boot loaded: 0xA5
	r := []byte{0x0}
	if _, err := cl.pmb.iostream.Read(r); err != nil {
		return fmt.Errorf("error reading chaos loader ready message: %v", err)
	}
	if r[0] != 0xA5 {
		return fmt.Errorf("unknown chaos loader ready message %X", r[0])
	}
	shortDelay()

	// We need to send one ping to activate loader.
	pong, err := cl.Ping()
	if err != nil {
		return fmt.Errorf("error sending first ping: %v", err)
	}
	if !pong {
		return fmt.Errorf("chaos didn't reply to the first ping")
	}
	log.Print("Chaos bootloader activated")
	return nil
}

func (cl *ChaosLoader) Ping() (bool, error) {
	if _, err := cl.pmb.iostream.Write([]byte{'A'}); err != nil {
		return false, err
	}
	shortDelay()
	reply := []byte{0x00}
	if _, err := cl.pmb.iostream.Read(reply); err != nil {
		return false, err
	}
	if reply[0] == 'R' {
		return true, nil
	}
	return false, nil
}

// ReadInfo sends "Get info" command to the bootloader and dumps the result.
func (cl *ChaosLoader) ReadInfo() (ChaosPhoneInfo, error) {
	shortDelay()
	if _, err := cl.pmb.iostream.Write([]byte{'I'}); err != nil {
		return ChaosPhoneInfo{}, err
	}
	shortDelay()
	reply := make([]byte, 128)
	var n int
	var err error
	if n, err = cl.pmb.iostream.Read(reply); err != nil {
		return ChaosPhoneInfo{}, err
	}
	if n < 128 {
		return ChaosPhoneInfo{}, fmt.Errorf("less than 128 bytes read: %d", n)
	}

	info, err := ParseChaosInfo(bytes.NewBuffer(reply))
	if err != nil {
		return ChaosPhoneInfo{}, err
	}
	bm := info.BlockMap
	cl.bm = &bm
	return info, nil
}

func (cl *ChaosLoader) readAndCheck(maxN int) ([]byte, error) {
	// This is max what we could ever read, but the actual read amount will likely
	// be smaller.
	replyBuf := make([]byte, maxN)
	var n int
	var err error
	if n, err = cl.pmb.iostream.Read(replyBuf); err != nil {
		return nil, fmt.Errorf("cannot read flash: %v", err)
	}
	return replyBuf[:n], nil
}

// ReadFlash reads a memory region from Flash.
func (cl *ChaosLoader) ReadFlash(baseAddr int64, buf []byte) error {
	if cl.bm == nil {
		if _, err := cl.ReadInfo(); err != nil {
			return err
		}
	}

	reqLen := len(buf)

	// Construct the READ command.
	cmd := readCmd{
		Cmd:  'R',
		Addr: uint32(baseAddr),
		Size: uint32(reqLen),
	}

	cmdBuf := new(bytes.Buffer)
	if err := binary.Write(cmdBuf, binary.BigEndian, cmd); err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	shortDelay()
	if _, err := cl.pmb.iostream.Write(cmdBuf.Bytes()); err != nil {
		return err
	}
	shortDelay()

	// We need the total length + 4 bytes control data.
	stillNeedToRead := reqLen + 4
	inBuffer := make([]byte, 0, stillNeedToRead)
	for stillNeedToRead > 0 {
		gotData, err := cl.readAndCheck(stillNeedToRead)
		if err != nil {
			return err
		}
		inBuffer = append(inBuffer, gotData...)
		stillNeedToRead -= len(gotData)
	}
	if len(inBuffer) != reqLen+4 {
		return fmt.Errorf("wrong lengh of received data (got %d, want %d)", len(inBuffer), len(buf)+4)
	}

	// Verify that the control data contains OK and that the checksum is correct.
	n := len(inBuffer)
	okSign := inBuffer[n-4 : n-2]
	if !bytes.Equal(okSign, []byte{'O', 'K'}) {
		return fmt.Errorf("didn't successfully receive the block, ok=%v", okSign)
	}
	chkBytes := inBuffer[n-2 : n]
	wantChk := chkBytes[0] // This should be just one byte, because the wanted CHK is a byte-wise XOR.
	gotChk := byte(0)
	for i := 0; i < n-4; i++ {
		gotChk ^= inBuffer[i]
	}
	if gotChk != wantChk {
		return fmt.Errorf("checksum doesn't match: got %X, want %X", gotChk, wantChk)
	}

	// copy() copies only so much data that the dest buffer can accomodate.
	copy(buf, inBuffer)
	return nil
}

// ParseChaosInfo parses an info dump saved in a file into a structure.
// TODO: Accept an io.Reader, not a path to file; return error as the second return value.
func ParseChaosInfo(r io.Reader) (ChaosPhoneInfo, error) {

	var info chaosInfo
	if err := binary.Read(r, binary.LittleEndian, &info); err != nil {
		fmt.Println("failed to Read:", err)
		return ChaosPhoneInfo{}, err
	}

	phoneInfo := ChaosPhoneInfo{
		ModelName:    string(info.ModelName[:]),
		Manufacturer: string(info.Manufacturer[:]),
		IMEI:         string(info.IMEI[:]),
	}

	phoneInfo.BlockMap = blockman.New(int64(info.FlashBaseAddr))
	if info.FlashRegionsNum >= 1 {
		phoneInfo.BlockMap.AddRegion(int64(info.FlashRegion0BlockSizeDiv256)*256, int(info.FlashRegion0BlocksNumMinus1)+1)
	}
	if info.FlashRegionsNum >= 2 {
		phoneInfo.BlockMap.AddRegion(int64(info.FlashRegion1BlockSizeDiv256)*256, int(info.FlashRegion1BlocksNumMinus1)+1)
	}
	if info.FlashRegionsNum >= 3 {
		phoneInfo.BlockMap.AddRegion(int64(info.FlashRegion2BlockSizeDiv256)*256, int(info.FlashRegion2BlocksNumMinus1)+1)
	}
	if info.FlashRegionsNum >= 4 {
		phoneInfo.BlockMap.AddRegion(int64(info.FlashRegion3BlockSizeDiv256)*256, int(info.FlashRegion3BlocksNumMinus1)+1)
	}
	if info.FlashRegionsNum > 4 {
		return ChaosPhoneInfo{}, fmt.Errorf("unsupported number of regions: %d", info.FlashRegionsNum)
	}

	return phoneInfo, nil
}
