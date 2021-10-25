package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/api"
)

func motionTab(parent fyne.Window, client *api.Client) *fyne.Container {
	// Create label for heart rate
	xCoordLbl := newText("0", 24)
	// Creae container to store heart rate section
	xCoordSect := container.NewVBox(
		newText("X Coordinate", 12),
		xCoordLbl,
		canvas.NewLine(theme.ShadowColor()),
	)

	// Create label for heart rate
	yCoordLbl := newText("0", 24)
	// Creae container to store heart rate section
	yCoordSect := container.NewVBox(
		newText("Y Coordinate", 12),
		yCoordLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	// Create label for heart rate
	zCoordLbl := newText("0", 24)
	// Creae container to store heart rate section
	zCoordSect := container.NewVBox(
		newText("Z Coordinate", 12),
		zCoordLbl,
		canvas.NewLine(theme.ShadowColor()),
	)

	// Create variable to keep track of whether motion started
	started := false

	// Create button to stop motion
	stopBtn := widget.NewButton("Stop", nil)
	// Create button to start motion
	startBtn := widget.NewButton("Start", func() {
		// if motion is started
		if started {
			// Do nothing
			return
		}
		// Set motion started
		started = true
		// Watch motion values
		motionCh, cancel, err := client.WatchMotion()
		if err != nil {
			guiErr(err, "Error getting heart rate channel", true, parent)
		}
		// Create done channel
		done := make(chan struct{}, 1)
		go func() {
			for {
				select {
				case <-done:
					return
				case motion := <-motionCh:
					// Set labels to new values
					xCoordLbl.Text = strconv.Itoa(int(motion.X))
					yCoordLbl.Text = strconv.Itoa(int(motion.Y))
					zCoordLbl.Text = strconv.Itoa(int(motion.Z))
					// Refresh labels to display new values
					xCoordLbl.Refresh()
					yCoordLbl.Refresh()
					zCoordLbl.Refresh()
				}
			}
		}()
		// Create stop function
		stopBtn.OnTapped = func() {
			done <- struct{}{}
			started = false
			cancel()
		}

	})
	// Run stop button function on close if possible
	onClose = append(onClose, func() {
		if stopBtn.OnTapped != nil {
			stopBtn.OnTapped()
		}
	})

	// Return new container containing all elements
	return container.NewVBox(
		// Add rectangle for a bit of padding
		canvas.NewRectangle(color.Transparent),
		startBtn,
		stopBtn,
		xCoordSect,
		yCoordSect,
		zCoordSect,
	)
}
