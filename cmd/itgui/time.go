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

func timeTab(parent fyne.Window, client *api.Client) *fyne.Container {
	// Create new entry for time string
	timeEntry := widget.NewEntry()
	// Set text to current time formatter properly
	timeEntry.SetText(time.Now().Format(time.RFC1123))

	// Create button to set current time
	currentBtn := widget.NewButton("Set Current", func() {
		timeEntry.SetText(time.Now().Format(time.RFC1123))
		setTime(client, true)
	})

	// Create button to set time inside entry
	timeBtn := widget.NewButton("Set", func() {
		// Parse time as RFC1123 string
		parsedTime, err := time.Parse(time.RFC1123, timeEntry.Text)
		if err != nil {
			guiErr(err, "Error parsing time string", false, parent)
			return
		}
		// Set time to parsed time
		setTime(client, false, parsedTime)
	})

	// Return new container with all elements centered
	return container.NewVBox(
		layout.NewSpacer(),
		timeEntry,
		currentBtn,
		timeBtn,
		layout.NewSpacer(),
	)
}

// setTime sets the first element in the variadic parameter
// if current is false, otherwise, it sets the current time.
func setTime(client *api.Client, current bool, t ...time.Time) error {
	var err error
	if current {
		err = client.SetTime(context.Background(), time.Now())
	} else {
		err = client.SetTime(context.Background(), t[0])
	}
	if err != nil {
		return err
	}
	return nil
}
