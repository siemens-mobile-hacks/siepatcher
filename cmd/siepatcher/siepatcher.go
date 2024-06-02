package main

import (
	"bytes"
	"log"

	_ "embed"

	"image/color"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/siemens-mobile-hacks/siepatcher/pkg/pmb887x"
)

type Event int

const (
	ConnectTarget Event = iota // 0
	TargetInfo
	CmdError
	CmdProgress
)

type ConnectInfoType struct {
	SerialPath    string
	SerialSpeed   string
	EmuSocketPath string
	FFPath        string
}

type PatcherCommand struct {
	EventType   Event
	ConnectInfo ConnectInfoType
}

type PatcherReply struct {
	EventType  Event
	DeviceInfo struct {
		PhoneInfo pmb887x.ChaosPhoneInfo
	}
	ProgressDescr string
	ErrorDescr    string
}

var (
	//go:embed bin/phoneimage.jpeg
	headerImageBin []byte
)

func main() {
	patcherApp := app.NewWithID("com.kibab.siepatcher")
	mainWin := patcherApp.NewWindow("SiePatcher")

	patcherCommands := make(chan PatcherCommand)
	patcherReplies := make(chan PatcherReply)

	// Real device connection settings: serial port config.
	serialName := widget.NewEntry()
	serialName.SetPlaceHolder("Serial port name/path...")

	serialSpeed := widget.NewSelect([]string{"115200", "230400", "460800", "614400", "921600", "1228800", "1600000", "1500000", "1625000", "3250000"}, nil)
	serialConfig := container.New(layout.NewFormLayout(), widget.NewLabel("Serial port:"), serialName, widget.NewLabel("Speed:"), serialSpeed)

	// Emulator settings: path to socket.
	emuSocketPath := widget.NewEntry()
	emuSocketPath.SetPlaceHolder("Path to emulator socket...")
	emuConfig := container.NewVBox(widget.NewLabel("Socket path:"), emuSocketPath)

	// Fullflash file: path to file.
	ffFilePath := widget.NewEntry()
	ffFilePath.SetPlaceHolder("Path to fullflash file...")
	ffConfig := container.NewVBox(widget.NewLabel("Fullflash path:"), ffFilePath)

	// Configure tabs for selecting a target.
	realDeviceTab := container.NewTabItem("Real device", serialConfig)
	emuTab := container.NewTabItem("Emulator", emuConfig)
	ffTab := container.NewTabItem("Fullflash", ffConfig)
	targetConfig := container.NewAppTabs(
		realDeviceTab,
		emuTab,
		ffTab,
	)

	infoBox := widget.NewEntry()
	infoBox.MultiLine = true
	infoBox.Disable()
	infoBox.SetMinRowsVisible(10)
	infoBox.SetPlaceHolder("Press 'Connect'...")

	// An image in the header.
	img := canvas.NewImageFromReader(bytes.NewReader(headerImageBin), "phone.jpg")
	img.FillMode = canvas.ImageFillOriginal

	// A status bar.
	statusIcon := widget.NewIcon(theme.MediaRecordIcon())
	statusText := canvas.NewText("Offline", color.RGBA{0xFF, 00, 00, 0xFF})
	statusBar := container.NewHBox(statusIcon, statusText)

	content := container.NewVBox(img, targetConfig, widget.NewButton("Connect and get info", func() {
		patcherApp.Preferences().SetString("serial_path", serialName.Text)
		patcherApp.Preferences().SetString("serial_speed", serialSpeed.Selected)
		patcherApp.Preferences().SetString("emu_socket_path", emuSocketPath.Text)
		patcherApp.Preferences().SetString("ff_file_path", ffFilePath.Text)

		switch targetConfig.Selected() {
		case realDeviceTab:
			log.Printf("Using real device @ %q speed %s", serialName.Text, serialSpeed.Selected)
			cmd := PatcherCommand{
				EventType: ConnectTarget,
				ConnectInfo: ConnectInfoType{
					SerialPath:  serialName.Text,
					SerialSpeed: serialSpeed.Selected,
				},
			}
			patcherCommands <- cmd
		case emuTab:
			log.Printf("Using emulator @ socket path %q", emuSocketPath.Text)
		case ffTab:
			log.Printf("Using fullflash file @ path %q", ffFilePath.Text)
		}
	}), infoBox, statusBar)

	// Load preferences.
	log.Printf("Serial from settings: %s", patcherApp.Preferences().String("serial_path"))
	log.Printf("Speed from settings: %s", patcherApp.Preferences().String("serial_speed"))
	serialName.Text = patcherApp.Preferences().String("serial_path")
	serialSpeed.SetSelected(patcherApp.Preferences().String("serial_speed"))
	emuSocketPath.Text = patcherApp.Preferences().String("emu_socket_path")
	ffFilePath.Text = patcherApp.Preferences().String("ff_file_path")

	mainWin.SetContent(content)

	// Start a goroutine to react to commands from the UI.
	// The patcher will run there.
	go PatchEngine(patcherCommands, patcherReplies)

	// Start a goroutine to react to the events from the patcher.
	// We will update GUI here.
	go func(ch <-chan PatcherReply) {
		for ev := range ch {
			log.Printf("Callback: Got a reply %v", ev)
			switch ev.EventType {
			case TargetInfo:
				infoBox.SetText(ev.DeviceInfo.PhoneInfo.String())
				statusText.Color = color.RGBA{0, 255, 0, 255}
				statusText.Text = "Online"
				statusBar.Refresh()
			case CmdError:
				infoBox.SetText(ev.ErrorDescr)
			case CmdProgress:
				log.Printf("Got a progress report: progress = %s", ev.ProgressDescr)
				statusText.Color = color.RGBA{255, 168, 0, 255}
				statusText.Text = ev.ProgressDescr
				statusBar.Refresh()
			}
		}
	}(patcherReplies)

	mainWin.ShowAndRun()
}
