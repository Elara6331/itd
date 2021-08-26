package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/mitchellh/mapstructure"
	"go.arsenm.dev/itd/internal/types"
)

func upgradeTab(parent fyne.Window) *fyne.Container {
	var (
		archivePath string
		fiwmarePath string
		initPktPath string
	)

	// Create archive selection dialog
	archiveDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		archivePath = uc.URI().Path()
	}, parent)
	// Limit dialog to .zip files
	archiveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".zip"}))
	// Create button to show dialog
	archiveBtn := widget.NewButton("Select archive (.zip)", archiveDialog.Show)

	// Create firmware selection dialog
	firmwareDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		fiwmarePath = uc.URI().Path()
	}, parent)
	// Limit dialog to .bin files
	firmwareDialog.SetFilter(storage.NewExtensionFileFilter([]string{".bin"}))
	// Create button to show dialog
	firmwareBtn := widget.NewButton("Select init packet (.bin)", firmwareDialog.Show)

	// Create init packet selection dialog
	initPktDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		initPktPath = uc.URI().Path()
	}, parent)
	// Limit dialog to .dat files
	initPktDialog.SetFilter(storage.NewExtensionFileFilter([]string{".dat"}))
	// Create button to show dialog
	initPktBtn := widget.NewButton("Select init packet (.dat)", initPktDialog.Show)

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
		if archivePath == "" && (initPktPath == "" && fiwmarePath == "") {
			guiErr(nil, "Upgrade requires archive or files selected", parent)
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

		var fwUpgType int
		var files []string
		// Get appropriate upgrade type and file paths
		switch upgradeTypeSelect.Selected {
		case "Archive":
			fwUpgType = types.UpgradeTypeArchive
			files = append(files, archivePath)
		case "Files":
			fwUpgType = types.UpgradeTypeFiles
			files = append(files, initPktPath, fiwmarePath)
		}

		// Dial itd UNIX socket
		conn, err := net.Dial("unix", SockPath)
		if err != nil {
			guiErr(err, "Error dialing socket", parent)
			return
		}
		defer conn.Close()

		// Encode firmware upgrade request to connection
		json.NewEncoder(conn).Encode(types.Request{
			Type: types.ReqTypeFwUpgrade,
			Data: types.ReqDataFwUpgrade{
				Type:  fwUpgType,
				Files: files,
			},
		})

		// Show progress dialog
		progressDlg.Show()
		// Hide progress dialog after completion
		defer progressDlg.Hide()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var res types.Response
			// Decode scanned line into response struct
			err = json.Unmarshal(scanner.Bytes(), &res)
			if err != nil {
				guiErr(err, "Error decoding response", parent)
				return
			}
			if res.Error {
				guiErr(err, "Error returned in response", parent)
				return
			}
			var event types.DFUProgress
			// Decode response data into progress struct
			err = mapstructure.Decode(res.Value, &event)
			if err != nil {
				guiErr(err, "Error decoding response value", parent)
				return
			}
			// If transfer finished, break
			if event.Received == event.Total {
				break
			}
			// Set label text to received / total B
			progressLbl.SetText(fmt.Sprintf("%d / %d B", event.Received, event.Total))
			// Set progress bar values
			progressBar.Max = float64(event.Total)
			progressBar.Value = float64(event.Received)
			// Refresh progress bar
			progressBar.Refresh()
		}
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
