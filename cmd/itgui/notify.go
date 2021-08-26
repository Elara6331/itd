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
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Title")

	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetPlaceHolder("Body")

	sendBtn := widget.NewButton("Send", func() {
		conn, err := net.Dial("unix", SockPath)
		if err != nil {
			guiErr(err, "Error dialing socket", parent)
			return
		}
		json.NewEncoder(conn).Encode(types.Request{
			Type: types.ReqTypeNotify,
			Data: types.ReqDataNotify{
				Title: titleEntry.Text,
				Body:  bodyEntry.Text,
			},
		})
	})

	return container.NewVBox(
		layout.NewSpacer(),
		titleEntry,
		bodyEntry,
		sendBtn,
		layout.NewSpacer(),
	)
}
