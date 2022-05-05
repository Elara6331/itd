package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/api"
)

func notifyTab(ctx context.Context, client *api.Client, w fyne.Window) fyne.CanvasObject {
	c := container.NewVBox()
	c.Add(layout.NewSpacer())

	// Create new entry for title
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Title")
	c.Add(titleEntry)

	// Create new multiline entry for body
	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetPlaceHolder("Body")
	c.Add(bodyEntry)

	// Create new send button
	sendBtn := widget.NewButton("Send", func() {
		// Send notification
		err := client.Notify(ctx, titleEntry.Text, bodyEntry.Text)
		if err != nil {
			guiErr(err, "Error sending notification", false, w)
			return
		}
	})
	c.Add(sendBtn)

	c.Add(layout.NewSpacer())
	return container.NewVScroll(c)
}
