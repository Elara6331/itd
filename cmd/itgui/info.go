package main

import (
	"errors"
	"fmt"
	"image/color"

	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"go.arsenm.dev/itd/api"
	"go.arsenm.dev/itd/internal/types"
)

func infoTab(parent fyne.Window, client *api.Client) *fyne.Container {
	infoLayout := container.NewVBox(
		// Add rectangle for a bit of padding
		canvas.NewRectangle(color.Transparent),
	)

	// Create label for heart rate
	heartRateLbl := newText("0 BPM", 24)
	// Creae container to store heart rate section
	heartRate := container.NewVBox(
		newText("Heart Rate", 12),
		heartRateLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(heartRate)

	fmt.Println(3)
	heartRateCh, cancel, err := client.WatchHeartRate()
	onClose = append(onClose, cancel)
	go func() {
		for heartRate := range heartRateCh {
			// Change text of heart rate label
			heartRateLbl.Text = fmt.Sprintf("%d BPM", heartRate)
			// Refresh label
			heartRateLbl.Refresh()
		}
	}()
	fmt.Println(4)

	// Create label for battery level
	battLevelLbl := newText("0%", 24)
	// Create container to store battery level section
	battLevel := container.NewVBox(
		newText("Battery Level", 12),
		battLevelLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(battLevel)

	fmt.Println(5)
	battLevelCh, cancel, err := client.WatchBatteryLevel()
	onClose = append(onClose, cancel)
	go func() {
		for battLevel := range battLevelCh {
			// Change text of battery level label
			battLevelLbl.Text = fmt.Sprintf("%d%%", battLevel)
			// Refresh label
			battLevelLbl.Refresh()
		}
	}()
	fmt.Println(6)

	fmt.Println(7)
	fwVerString, err := client.Version()
	if err != nil {
		guiErr(err, "Error getting firmware string", true, parent)
	}
	fmt.Println(8)

	fwVer := container.NewVBox(
		newText("Firmware Version", 12),
		newText(fwVerString, 24),
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(fwVer)

	fmt.Println(9)
	btAddrString, err := client.Address()
	if err != nil {
		panic(err)
	}
	fmt.Println(10)

	btAddr := container.NewVBox(
		newText("Bluetooth Address", 12),
		newText(btAddrString, 24),
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(btAddr)

	return infoLayout
}

func getResp(line []byte) (*types.Response, error) {
	var res types.Response
	err := json.Unmarshal(line, &res)
	if err != nil {
		return nil, err
	}
	if res.Error {
		return nil, errors.New(res.Message)
	}
	return &res, nil
}

func newText(t string, size float32) *canvas.Text {
	text := canvas.NewText(t, theme.ForegroundColor())
	text.TextSize = size
	return text
}
