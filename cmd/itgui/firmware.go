package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"go.elara.ws/itd/api"
)

func firmwareTab(ctx context.Context, client *api.Client, w fyne.Window) fyne.CanvasObject {
	// Create select to chose between archive and files upgrade
	typeSelect := widget.NewSelect([]string{"Archive", "Files"}, nil)
	typeSelect.PlaceHolder = "Upgrade Type"

	// Create map to store files
	files := map[string]string{}

	// Create and disable start button
	startBtn := widget.NewButton("Start", nil)
	startBtn.Disable()

	// Create new file open dialog for archive
	archiveDlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()
		// Set archive path in map
		files[".zip"] = uc.URI().Path()
		// Enable start button
		startBtn.Enable()
	}, w)
	// Only allow .zip files
	archiveDlg.SetFilter(storage.NewExtensionFileFilter([]string{".zip"}))
	// Create button to show dialog
	archiveBtn := widget.NewButton("Select Archive (.zip)", archiveDlg.Show)

	// Create new file open dialog for firmware image
	imageDlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()

		// Set firmware image path in map
		files[".bin"] = uc.URI().Path()

		// If the init packet was already selected
		_, datOk := files[".dat"]
		if datOk {
			// Enable start button
			startBtn.Enable()
		}
	}, w)
	// Only allow .bin files
	imageDlg.SetFilter(storage.NewExtensionFileFilter([]string{".bin"}))
	// Create button to show dialog
	imageBtn := widget.NewButton("Select Firmware (.bin)", imageDlg.Show)

	// Create new file open dialog for init packet
	initDlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()

		// Set init packet path in map
		files[".dat"] = uc.URI().Path()

		// If the firmware image was already selected
		_, binOk := files[".bin"]
		if binOk {
			// Enable start button
			startBtn.Enable()
		}
	}, w)
	// Only allow .dat files
	initDlg.SetFilter(storage.NewExtensionFileFilter([]string{".dat"}))
	// Create button to show dialog
	initBtn := widget.NewButton("Select Init Packet (.dat)", initDlg.Show)

	var upgType api.UpgradeType = 255
	// When upgrade type changes
	typeSelect.OnChanged = func(s string) {
		// Delete all files from map
		delete(files, ".bin")
		delete(files, ".dat")
		delete(files, ".zip")
		// Hide all dialog buttons
		imageBtn.Hide()
		initBtn.Hide()
		archiveBtn.Hide()
		// Disable start button
		startBtn.Disable()

		switch s {
		case "Files":
			// Set file upgrade type
			upgType = api.UpgradeTypeFiles
			// Show firmware image and init packet buttons
			imageBtn.Show()
			initBtn.Show()
		case "Archive":
			// Set archive upgrade type
			upgType = api.UpgradeTypeArchive
			// Show archive button
			archiveBtn.Show()
		}
	}
	// Select archive by default
	typeSelect.SetSelectedIndex(0)

	// When start button pressed
	startBtn.OnTapped = func() {
		var args []string
		// Append the appropriate files for upgrade type
		switch upgType {
		case api.UpgradeTypeArchive:
			args = append(args, files[".zip"])
		case api.UpgradeTypeFiles:
			args = append(args, files[".dat"], files[".bin"])
		}

		// If args are nil (invalid upgrade type)
		if args == nil {
			return
		}

		// Create new progress dialog
		progress := newProgress(w)
		// Start firmware upgrade
		progressCh, err := client.FirmwareUpgrade(ctx, upgType, args...)
		if err != nil {
			guiErr(err, "Error performing firmware upgrade", false, w)
			return
		}
		// Show progress dialog
		progress.Show()
		// For every progress event
		for progressEvt := range progressCh {
			// Set progress bar values
			progress.SetTotal(float64(progressEvt.Total))
			progress.SetValue(float64(progressEvt.Sent))
		}
		// Hide progress dialog
		progress.Hide()
	}

	return container.NewVBox(
		layout.NewSpacer(),
		typeSelect,
		archiveBtn,
		imageBtn,
		initBtn,
		startBtn,
		layout.NewSpacer(),
	)
}
