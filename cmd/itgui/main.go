package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

var SockPath = "/tmp/itd/socket"

func main() {
	a := app.New()
	window := a.NewWindow("itgui")

	tabs := container.NewAppTabs(
		container.NewTabItem("Info", infoTab(window)),
		container.NewTabItem("Notify", notifyTab(window)),
		container.NewTabItem("Set Time", timeTab(window)),
		container.NewTabItem("Upgrade", upgradeTab(window)),
	)

	window.SetContent(tabs)
	window.ShowAndRun()
}
