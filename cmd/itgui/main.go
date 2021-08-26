package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

var SockPath = "/tmp/itd/socket"

func main() {
	// Create new app
	a := app.New()
	// Create new window with title "itgui"
	window := a.NewWindow("itgui")

	// Create new app tabs container
	tabs := container.NewAppTabs(
		container.NewTabItem("Info", infoTab(window)),
		container.NewTabItem("Notify", notifyTab(window)),
		container.NewTabItem("Set Time", timeTab(window)),
		container.NewTabItem("Upgrade", upgradeTab(window)),
	)

	// Set tabs as window content
	window.SetContent(tabs)
	// Show window and run app
	window.ShowAndRun()
}