package main

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/api"
)

func motionTab(ctx context.Context, client *api.Client, w fyne.Window) fyne.CanvasObject {
	// Create titledText for each coordinate
	xText := newTitledText("X Coordinate", "0")
	yText := newTitledText("Y Coordinate", "0")
	zText := newTitledText("Z Coordinate", "0")

	var ctxCancel func()

	// Create start button
	toggleBtn := widget.NewButton("Start", nil)
	// Set button's on tapped callback
	toggleBtn.OnTapped = func() {
		switch toggleBtn.Text {
		case "Start":
			// Create new context for motion
			motionCtx, cancel := context.WithCancel(ctx)
			// Set ctxCancel to function so that stop button can run it
			ctxCancel = cancel
			// Watch motion
			motionCh, err := client.WatchMotion(motionCtx)
			if err != nil {
				guiErr(err, "Error watching motion", false, w)
				return
			}
			go func() {
				// For every motion event
				for motion := range motionCh {
					// Set coordinates
					xText.SetBody(fmt.Sprint(motion.X))
					yText.SetBody(fmt.Sprint(motion.Y))
					zText.SetBody(fmt.Sprint(motion.Z))
				}
			}()
			// Set button text to "Stop"
			toggleBtn.SetText("Stop")
		case "Stop":
			// Cancel motion context
			ctxCancel()
			// Set button text to "Start"
			toggleBtn.SetText("Start")
		}
	}

	return container.NewVScroll(container.NewVBox(
		toggleBtn,
		xText,
		yText,
		zText,
	))
}
