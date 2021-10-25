package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"go.arsenm.dev/itd/api"
)

var onClose []func()

func main() {
	// Create new app
	a := app.New()
	// Create new window with title "itgui"
	window := a.NewWindow("itgui")
	window.SetOnClosed(func() {
		for _, closeFn := range onClose {
			closeFn()
		}
	})

	client, err := api.New(api.DefaultAddr)
	if err != nil {
		guiErr(err, "Error connecting to itd", true, window)
	}
	onClose = append(onClose, func() {
		client.Close()
	})

	// Create new app tabs container
	tabs := container.NewAppTabs(
		container.NewTabItem("Info", infoTab(window, client)),
		container.NewTabItem("Notify", notifyTab(window, client)),
		container.NewTabItem("Set Time", timeTab(window, client)),
		container.NewTabItem("Upgrade", upgradeTab(window, client)),
	)

	// Set tabs as window content
	window.SetContent(tabs)
	// Show window and run app
	window.ShowAndRun()
}
