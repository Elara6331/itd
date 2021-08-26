package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func guiErr(err error, msg string, parent fyne.Window) {
	msgLbl := widget.NewLabel(msg)
	msgLbl.Wrapping = fyne.TextWrapWord
	msgLbl.Alignment = fyne.TextAlignCenter
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(350, 0))
	content := container.NewVBox(
		msgLbl,
		rect,
	)
	if err != nil {
		errLbl := widget.NewLabel(err.Error())
		content.Add(widget.NewAccordion(
			widget.NewAccordionItem("More Details", errLbl),
		))
	}
	dialog.NewCustom("Error", "Ok", content, parent).Show()
}
