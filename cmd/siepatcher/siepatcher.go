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
)

var (
	//go:embed bin/phoneimage.jpeg
	headerImageBin []byte
)

func main() {
	patcherApp := app.NewWithID("com.kibab.siepatcher")
	mainWin := patcherApp.NewWindow("SiePatcher")

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
	infoBox.SetMinRowsVisible(10)
	infoBox.SetPlaceHolder("Press 'Connect'...")

	// An image in the header.
	img := canvas.NewImageFromReader(bytes.NewReader(headerImageBin), "phone.jpg")
	img.FillMode = canvas.ImageFillOriginal

	// A status bar.
	statusIcon := widget.NewIcon(theme.MediaRecordIcon())
	statusText := canvas.NewText("Not connected", color.RGBA{0xFF, 00, 00, 0xFF})
	statusBar := container.NewHBox(statusIcon, statusText)

	content := container.NewVBox(img, targetConfig, widget.NewButton("Connect and get info", func() {
		patcherApp.Preferences().SetString("serial_path", serialName.Text)
		patcherApp.Preferences().SetString("serial_speed", serialSpeed.Selected)
		patcherApp.Preferences().SetString("emu_socket_path", emuSocketPath.Text)
		patcherApp.Preferences().SetString("ff_file_path", ffFilePath.Text)

		switch targetConfig.Selected() {
		case realDeviceTab:
			log.Printf("Using real device @ %q speed %s", serialName.Text, serialSpeed.Selected)
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

	mainWin.ShowAndRun()
}
