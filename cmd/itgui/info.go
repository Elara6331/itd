package main

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"go.arsenm.dev/itd/api"
)

func infoTab(ctx context.Context, client *api.Client, w fyne.Window) fyne.CanvasObject {
	c := container.NewVBox()

	// Create titled text for heart rate
	heartRateText := newTitledText("Heart Rate", "0 BPM")
	c.Add(heartRateText)
	// Watch heart rate
	heartRateCh, err := client.WatchHeartRate(ctx)
	if err != nil {
		guiErr(err, "Error watching heart rate", true, w)
	}
	go func() {
		// For every heart rate sample
		for heartRate := range heartRateCh {
			// Set body of titled text
			heartRateText.SetBody(fmt.Sprintf("%d BPM", heartRate))
		}
	}()

	// Create titled text for battery level
	battLevelText := newTitledText("Battery Level", "0%")
	c.Add(battLevelText)
	// Watch battery level
	battLevelCh, err := client.WatchBatteryLevel(ctx)
	if err != nil {
		guiErr(err, "Error watching battery level", true, w)
	}
	go func() {
		// For every battery level sample
		for battLevel := range battLevelCh {
			// Set body of titled text
			battLevelText.SetBody(fmt.Sprintf("%d%%", battLevel))
		}
	}()

	// Create titled text for step count
	stepCountText := newTitledText("Step Count", "0 Steps")
	c.Add(stepCountText)
	// Watch step count
	stepCountCh, err := client.WatchStepCount(ctx)
	if err != nil {
		guiErr(err, "Error watching step count", true, w)
	}
	go func() {
		// For every step count sample
		for stepCount := range stepCountCh {
			// Set body of titled text
			stepCountText.SetBody(fmt.Sprintf("%d Steps", stepCount))
		}
	}()

	// Create new titled text for address
	addressText := newTitledText("Address", "")
	c.Add(addressText)
	// Get address
	address, err := client.Address(ctx)
	if err != nil {
		guiErr(err, "Error getting address", true, w)
	}
	// Set body of titled text
	addressText.SetBody(address)

	// Create new titled text for version
	versionText := newTitledText("Version", "")
	c.Add(versionText)
	// Get version
	version, err := client.Version(ctx)
	if err != nil {
		guiErr(err, "Error getting version", true, w)
	}
	// Set body of titled text
	versionText.SetBody(version)

	return container.NewVScroll(c)
}
