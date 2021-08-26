package main

import (
	"encoding/json"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/internal/types"
)

func notifyTab(parent fyne.Window) *fyne.Container {
	// Create new entry for notification title
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Title")

	// Create multiline entry for notification body
	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetPlaceHolder("Body")

	// Create new button to send notification
	sendBtn := widget.NewButton("Send", func() {
		// Dial itd UNIX socket
		conn, err := net.Dial("unix", SockPath)
		if err != nil {
			guiErr(err, "Error dialing socket", parent)
			return
		}
		// Encode notify request on connection
		json.NewEncoder(conn).Encode(types.Request{
			Type: types.ReqTypeNotify,
			Data: types.ReqDataNotify{
				Title: titleEntry.Text,
				Body:  bodyEntry.Text,
			},
		})
	})

	// Return new container containing all elements
	return container.NewVBox(
		layout.NewSpacer(),
		titleEntry,
		bodyEntry,
		sendBtn,
		layout.NewSpacer(),
	)
}
