package main

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/itd/api"
	"go.arsenm.dev/itd/internal/types"
)

func upgradeTab(parent fyne.Window, client *api.Client) *fyne.Container {
	var (
		archivePath  string
		firmwarePath string
		initPktPath  string
	)

	var archiveBtn *widget.Button
	// Create archive selection dialog
	archiveDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		archivePath = uc.URI().Path()
		archiveBtn.SetText(fmt.Sprintf("Select archive (.zip) [%s]", filepath.Base(archivePath)))
	}, parent)
	// Limit dialog to .zip files
	archiveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".zip"}))
	// Create button to show dialog
	archiveBtn = widget.NewButton("Select archive (.zip)", archiveDialog.Show)

	var firmwareBtn *widget.Button
	// Create firmware selection dialog
	firmwareDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		firmwarePath = uc.URI().Path()
		firmwareBtn.SetText(fmt.Sprintf("Select firmware (.bin) [%s]", filepath.Base(firmwarePath)))
	}, parent)
	// Limit dialog to .bin files
	firmwareDialog.SetFilter(storage.NewExtensionFileFilter([]string{".bin"}))
	// Create button to show dialog
	firmwareBtn = widget.NewButton("Select firmware (.bin)", firmwareDialog.Show)

	var initPktBtn *widget.Button
	// Create init packet selection dialog
	initPktDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		initPktPath = uc.URI().Path()
		initPktBtn.SetText(fmt.Sprintf("Select init packet (.dat) [%s]", filepath.Base(initPktPath)))
	}, parent)
	// Limit dialog to .dat files
	initPktDialog.SetFilter(storage.NewExtensionFileFilter([]string{".dat"}))
	// Create button to show dialog
	initPktBtn = widget.NewButton("Select init packet (.dat)", initPktDialog.Show)

	// Hide init packet and firmware buttons
	initPktBtn.Hide()
	firmwareBtn.Hide()

	// Create dropdown to select upgrade type
	upgradeTypeSelect := widget.NewSelect([]string{
		"Archive",
		"Files",
	}, func(s string) {
		// Hide all buttons
		archiveBtn.Hide()
		initPktBtn.Hide()
		firmwareBtn.Hide()
		// Unhide appropriate button(s)
		switch s {
		case "Archive":
			archiveBtn.Show()
		case "Files":
			initPktBtn.Show()
			firmwareBtn.Show()
		}
	})
	// Select first elemetn
	upgradeTypeSelect.SetSelectedIndex(0)

	// Create new button to start DFU
	startBtn := widget.NewButton("Start", func() {
		// If archive path does not exist and both init packet and firmware paths
		// also do not exist, return error
		if archivePath == "" && (initPktPath == "" && firmwarePath == "") {
			guiErr(nil, "Upgrade requires archive or files selected", false, parent)
			return
		}

		// Create new label for byte progress
		progressLbl := widget.NewLabelWithStyle("0 / 0 B", fyne.TextAlignCenter, fyne.TextStyle{})
		// Create new progress bar
		progressBar := widget.NewProgressBar()
		// Create modal dialog containing label and progress bar
		progressDlg := widget.NewModalPopUp(container.NewVBox(
			layout.NewSpacer(),
			progressLbl,
			progressBar,
			layout.NewSpacer(),
		), parent.Canvas())
		// Resize modal to 300x100
		progressDlg.Resize(fyne.NewSize(300, 100))

		var fwUpgType api.UpgradeType
		var files []string
		// Get appropriate upgrade type and file paths
		switch upgradeTypeSelect.Selected {
		case "Archive":
			fwUpgType = types.UpgradeTypeArchive
			files = append(files, archivePath)
		case "Files":
			fwUpgType = types.UpgradeTypeFiles
			files = append(files, initPktPath, firmwarePath)
		}

		progress, err := client.FirmwareUpgrade(fwUpgType, files...)
		if err != nil {
			guiErr(err, "Error initiating DFU", false, parent)
			return
		}

		// Show progress dialog
		progressDlg.Show()

		for event := range progress {
			// Set label text to received / total B
			progressLbl.SetText(fmt.Sprintf("%d / %d B", event.Received, event.Total))
			// Set progress bar values
			progressBar.Max = float64(event.Total)
			progressBar.Value = float64(event.Received)
			// Refresh progress bar
			progressBar.Refresh()
			// If transfer finished, break
			if event.Sent == event.Total {
				break
			}
		}

		// Hide progress dialog after completion
		progressDlg.Hide()

		// Reset screen to default
		upgradeTypeSelect.SetSelectedIndex(0)
		firmwareBtn.SetText("Select firmware (.bin)")
		initPktBtn.SetText("Select init packet (.dat)")
		archiveBtn.SetText("Select archive (.zip)")
		firmwarePath = ""
		initPktPath = ""
		archivePath = ""

		dialog.NewInformation(
			"Upgrade Complete",
			"The firmware was transferred successfully.\nRemember to validate the firmware in InfiniTime settings.",
			parent,
		).Show()
	})

	// Return container containing all elements
	return container.NewVBox(
		layout.NewSpacer(),
		upgradeTypeSelect,
		archiveBtn,
		firmwareBtn,
		initPktBtn,
		startBtn,
		layout.NewSpacer(),
	)
}
