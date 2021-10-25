package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"go.arsenm.dev/itd/api"
)

func infoTab(parent fyne.Window, client *api.Client) *fyne.Container {
	infoLayout := container.NewVBox(
		// Add rectangle for a bit of padding
		canvas.NewRectangle(color.Transparent),
	)

	// Create label for heart rate
	heartRateLbl := newText("0 BPM", 24)
	// Creae container to store heart rate section
	heartRateSect := container.NewVBox(
		newText("Heart Rate", 12),
		heartRateLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(heartRateSect)

	heartRateCh, cancel, err := client.WatchHeartRate()
	if err != nil {
		guiErr(err, "Error getting heart rate channel", true, parent)
	}
	onClose = append(onClose, cancel)
	go func() {
		for heartRate := range heartRateCh {
			// Change text of heart rate label
			heartRateLbl.Text = fmt.Sprintf("%d BPM", heartRate)
			// Refresh label
			heartRateLbl.Refresh()
		}
	}()

	// Create label for heart rate
	stepCountLbl := newText("0 Steps", 24)
	// Creae container to store heart rate section
	stepCountSect := container.NewVBox(
		newText("Step Count", 12),
		stepCountLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(stepCountSect)

	stepCountCh, cancel, err := client.WatchStepCount()
	if err != nil {
		guiErr(err, "Error getting step count channel", true, parent)
	}
	onClose = append(onClose, cancel)
	go func() {
		for stepCount := range stepCountCh {
			// Change text of heart rate label
			stepCountLbl.Text = fmt.Sprintf("%d Steps", stepCount)
			// Refresh label
			stepCountLbl.Refresh()
		}
	}()

	// Create label for battery level
	battLevelLbl := newText("0%", 24)
	// Create container to store battery level section
	battLevel := container.NewVBox(
		newText("Battery Level", 12),
		battLevelLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(battLevel)

	battLevelCh, cancel, err := client.WatchBatteryLevel()
	if err != nil {
		guiErr(err, "Error getting battery level channel", true, parent)
	}
	onClose = append(onClose, cancel)
	go func() {
		for battLevel := range battLevelCh {
			// Change text of battery level label
			battLevelLbl.Text = fmt.Sprintf("%d%%", battLevel)
			// Refresh label
			battLevelLbl.Refresh()
		}
	}()

	fwVerString, err := client.Version()
	if err != nil {
		guiErr(err, "Error getting firmware string", true, parent)
	}

	fwVer := container.NewVBox(
		newText("Firmware Version", 12),
		newText(fwVerString, 24),
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(fwVer)

	btAddrString, err := client.Address()
	if err != nil {
		panic(err)
	}

	btAddr := container.NewVBox(
		newText("Bluetooth Address", 12),
		newText(btAddrString, 24),
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(btAddr)

	return infoLayout
}

func newText(t string, size float32) *canvas.Text {
	text := canvas.NewText(t, theme.ForegroundColor())
	text.TextSize = size
	return text
}
