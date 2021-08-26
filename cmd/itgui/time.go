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

	timeEntry := widget.NewEntry()
	timeEntry.SetText(time.Now().Format(time.RFC1123))

	currentBtn := widget.NewButton("Set Current", func() {
		timeEntry.SetText(time.Now().Format(time.RFC1123))
		setTime(true)
	})

	timeBtn := widget.NewButton("Set", func() {
		parsedTime, err := time.Parse(time.RFC1123, timeEntry.Text)
		if err != nil {
			guiErr(err, "Error parsing time string", parent)
			return
		}
		setTime(false, parsedTime)
	})

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
	conn, err := net.Dial("unix", SockPath)
	if err != nil {
		return err
	}
	var data string
	if current {
		data = "now"
	} else {
		data = t[0].Format(time.RFC3339)
	}
	defer conn.Close()
	err = json.NewEncoder(conn).Encode(types.Request{
		Type: types.ReqTypeSetTime,
		Data: data,
	})
	if err != nil {
		return err
	}
	return nil
}
