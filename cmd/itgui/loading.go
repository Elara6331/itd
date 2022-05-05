package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newLoadingPopUp(w fyne.Window) *widget.PopUp {
	pb := widget.NewProgressBarInfinite()
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	return widget.NewModalPopUp(
		container.NewMax(rect, pb),
		w.Canvas(),
	)
}
