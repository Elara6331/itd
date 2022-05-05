package main

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/api"
)

func timeTab(ctx context.Context, client *api.Client, w fyne.Window) fyne.CanvasObject {
	c := container.NewVBox()
	c.Add(layout.NewSpacer())

	// Create entry for time string
	timeEntry := widget.NewEntry()
	timeEntry.SetText(time.Now().Format(time.RFC1123))
	timeEntry.SetPlaceHolder("RFC1123")

	// Create button to set current time
	setCurrentBtn := widget.NewButton("Set current time", func() {
		// Set current time
		err := client.SetTime(ctx, time.Now())
		if err != nil {
			guiErr(err, "Error setting time", false, w)
			return
		}
		// Set time entry to current time
		timeEntry.SetText(time.Now().Format(time.RFC1123))
	})

	// Create button to set time from entry
	setBtn := widget.NewButton("Set", func() {
		// Parse RFC1123 time string in entry
		newTime, err := time.Parse(time.RFC1123, timeEntry.Text)
		if err != nil {
			guiErr(err, "Error parsing time string", false, w)
			return
		}
		// Set time from parsed string
		err = client.SetTime(ctx, newTime)
		if err != nil {
			guiErr(err, "Error setting time", false, w)
			return
		}
	})

	c.Add(timeEntry)
	c.Add(setBtn)
	c.Add(setCurrentBtn)

	c.Add(layout.NewSpacer())
	return c
}
