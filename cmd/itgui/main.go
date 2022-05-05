package main

import (
	"context"
	"sync"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"go.arsenm.dev/itd/api"
)

func main() {
	a := app.New()
	w := a.NewWindow("itgui")

	// Create new context for use with the API client
	ctx, cancel := context.WithCancel(context.Background())

	// Connect to ITD API
	client, err := api.New(api.DefaultAddr)
	if err != nil {
		guiErr(err, "Error connecting to ITD", true, w)
	}

	// Create channel to signal that the fs tab has been opened
	fsOpened := make(chan struct{})
	fsOnce := &sync.Once{}

	// Create app tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Info", infoTab(ctx, client, w)),
		container.NewTabItem("Motion", motionTab(ctx, client, w)),
		container.NewTabItem("Notify", notifyTab(ctx, client, w)),
		container.NewTabItem("FS", fsTab(ctx, client, w, fsOpened)),
		container.NewTabItem("Time", timeTab(ctx, client, w)),
		container.NewTabItem("Firmware", firmwareTab(ctx, client, w)),
	)

	// When a tab is selected
	tabs.OnSelected = func(ti *container.TabItem) {
		// If the tab's name is FS
		if ti.Text == "FS" {
			// Signal fsOpened only once
			fsOnce.Do(func() {
				fsOpened <- struct{}{}
			})
		}
	}

	// Cancel context on close
	w.SetOnClosed(cancel)
	// Set content and show window
	w.SetContent(tabs)
	w.ShowAndRun()
}
