package main

import (
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func guiErr(err error, msg string, fatal bool, parent fyne.Window) {
	// Create new label containing message
	msgLbl := widget.NewLabel(msg)
	// Text formatting settings
	msgLbl.Wrapping = fyne.TextWrapWord
	msgLbl.Alignment = fyne.TextAlignCenter
	// Create new rectangle to set the size of the dialog
	rect := canvas.NewRectangle(color.Transparent)
	// Set minimum size of rectangle to 350x0
	rect.SetMinSize(fyne.NewSize(350, 0))
	// Create new container containing message and rectangle
	content := container.NewVBox(
		msgLbl,
		rect,
	)
	if err != nil {
		// Create new label containing error text
		errEntry := widget.NewEntry()
		errEntry.SetText(err.Error())
		// If text changed, change it back
		errEntry.OnChanged = func(string) {
			errEntry.SetText(err.Error())
		}
		// Create new dropdown containing error label
		content.Add(widget.NewAccordion(
			widget.NewAccordionItem("More Details", errEntry),
		))
	}
	if fatal {
		// Create new error dialog
		errDlg := dialog.NewCustom("Error", "Close", content, parent)
		// On close, exit with code 1
		errDlg.SetOnClosed(func() {
			os.Exit(1)
		})
		// Show dialog
		errDlg.Show()
		// Run app prematurely to stop further execution
		parent.ShowAndRun()
	} else {
		// Show error dialog
		dialog.NewCustom("Error", "Ok", content, parent).Show()
	}
}
