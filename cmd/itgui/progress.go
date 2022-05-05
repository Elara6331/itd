package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type progress struct {
	lbl *widget.Label
	pb  *widget.ProgressBar
	*widget.PopUp
}

func newProgress(w fyne.Window) progress {
	out := progress{}

	// Create label to show how many bytes transfered and center it
	out.lbl = widget.NewLabel("0 / 0 B")
	out.lbl.Alignment = fyne.TextAlignCenter

	// Create new progress bar
	out.pb = widget.NewProgressBar()

	// Create new rectangle to set the size of the popup
	sizeRect := canvas.NewRectangle(color.Transparent)
	sizeRect.SetMinSize(fyne.NewSize(300, 50))
	
	// Create vbox for label and progress bar
	l := container.NewVBox(out.lbl, out.pb)
	// Create popup
	out.PopUp = widget.NewModalPopUp(container.NewMax(l, sizeRect), w.Canvas())

	return out
}

func (p progress) SetTotal(v float64) {
	p.pb.Max = v
	p.pb.Refresh()
	p.lbl.SetText(fmt.Sprintf("%.0f / %.0f B", p.pb.Value, v))
}

func (p progress) SetValue(v float64) {
	p.pb.SetValue(v)
	p.lbl.SetText(fmt.Sprintf("%.0f / %.0f B", v, p.pb.Max))
}
