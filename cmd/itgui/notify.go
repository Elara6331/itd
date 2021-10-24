package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/api"
)

func notifyTab(parent fyne.Window, client *api.Client) *fyne.Container {
	// Create new entry for notification title
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Title")

	// Create multiline entry for notification body
	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetPlaceHolder("Body")

	// Create new button to send notification
	sendBtn := widget.NewButton("Send", func() {
		err := client.Notify(titleEntry.Text, bodyEntry.Text)
		if err != nil {
			guiErr(err, "Error sending notification", false, parent)
			return
		}
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
