package main

import (
	"encoding/json"
	"net"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/internal/types"
)

func timeTab(parent fyne.Window) *fyne.Container {
	// Create new entry for time string
	timeEntry := widget.NewEntry()
	// Set text to current time formatter properly
	timeEntry.SetText(time.Now().Format(time.RFC1123))

	// Create button to set current time
	currentBtn := widget.NewButton("Set Current", func() {
		timeEntry.SetText(time.Now().Format(time.RFC1123))
		setTime(true)
	})

	// Create button to set time inside entry
	timeBtn := widget.NewButton("Set", func() {
		// Parse time as RFC1123 string
		parsedTime, err := time.Parse(time.RFC1123, timeEntry.Text)
		if err != nil {
			guiErr(err, "Error parsing time string", parent)
			return
		}
		// Set time to parsed time
		setTime(false, parsedTime)
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
func setTime(current bool, t ...time.Time) error {
	// Dial UNIX socket
	conn, err := net.Dial("unix", SockPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	var data string
	// If current is true, use the string "now"
	// otherwise, use the formatted time from the
	// first element in the variadic parameter.
	// "now" is more accurate than formatting
	// current time as only seconds are preserved
	// in that case.
	if current {
		data = "now"
	} else {
		data = t[0].Format(time.RFC3339)
	}
	// Encode SetTime request with above data
	err = json.NewEncoder(conn).Encode(types.Request{
		Type: types.ReqTypeSetTime,
		Data: data,
	})
	if err != nil {
		return err
	}
	return nil
}
