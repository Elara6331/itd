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

	archiveDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		archivePath = uc.URI().Path()
	}, parent)
	archiveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".zip"}))
	archiveBtn := widget.NewButton("Select archive (.zip)", archiveDialog.Show)

	firmwareDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		fiwmarePath = uc.URI().Path()
	}, parent)
	firmwareDialog.SetFilter(storage.NewExtensionFileFilter([]string{".bin"}))
	firmwareBtn := widget.NewButton("Select init packet (.bin)", firmwareDialog.Show)

	initPktDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if e != nil || uc == nil {
			return
		}
		uc.Close()
		initPktPath = uc.URI().Path()
	}, parent)
	initPktDialog.SetFilter(storage.NewExtensionFileFilter([]string{".dat"}))
	initPktBtn := widget.NewButton("Select init packet (.dat)", initPktDialog.Show)

	initPktBtn.Hide()
	firmwareBtn.Hide()

	upgradeTypeSelect := widget.NewSelect([]string{
		"Archive",
		"Files",
	}, func(s string) {
		archiveBtn.Hide()
		initPktBtn.Hide()
		firmwareBtn.Hide()
		switch s {
		case "Archive":
			archiveBtn.Show()
		case "Files":
			initPktBtn.Show()
			firmwareBtn.Show()
		}
	})
	upgradeTypeSelect.SetSelectedIndex(0)

	startBtn := widget.NewButton("Start", func() {
		if archivePath == "" && (initPktPath == "" && fiwmarePath == "") {
			guiErr(nil, "Upgrade requires archive or files selected", parent)
			return
		}

		progressLbl := widget.NewLabelWithStyle("0 / 0 B", fyne.TextAlignCenter, fyne.TextStyle{})
		progressBar := widget.NewProgressBar()
		progressDlg := widget.NewModalPopUp(container.NewVBox(
			layout.NewSpacer(),
			progressLbl,
			progressBar,
			layout.NewSpacer(),
		), parent.Canvas())
		progressDlg.Resize(fyne.NewSize(300, 100))

		var fwUpgType int
		var files []string
		switch upgradeTypeSelect.Selected {
		case "Archive":
			fwUpgType = types.UpgradeTypeArchive
			files = append(files, archivePath)
		case "Files":
			fwUpgType = types.UpgradeTypeFiles
			files = append(files, initPktPath, fiwmarePath)
		}

		conn, err := net.Dial("unix", SockPath)
		if err != nil {
			guiErr(err, "Error dialing socket", parent)
			return
		}
		defer conn.Close()

		json.NewEncoder(conn).Encode(types.Request{
			Type: types.ReqTypeFwUpgrade,
			Data: types.ReqDataFwUpgrade{
				Type:  fwUpgType,
				Files: files,
			},
		})

		progressDlg.Show()

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
			progressLbl.SetText(fmt.Sprintf("%d / %d B", event.Received, event.Total))
			progressBar.Max = float64(event.Total)
			progressBar.Value = float64(event.Received)
			progressBar.Refresh()
		}
		conn.Close()

		progressDlg.Hide()
	})

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
