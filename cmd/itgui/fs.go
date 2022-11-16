package main

import (
	"context"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/api"
)

func fsTab(ctx context.Context, client *api.Client, w fyne.Window, opened chan struct{}) fyne.CanvasObject {
	c := container.NewVBox()

	// Create new binding to store current directory
	cwdData := binding.NewString()
	cwdData.Set("/")

	// Create new list binding to store fs listing entries
	lsData := binding.NewUntypedList()

	// This goroutine waits until the fs tab is opened to
	// request the listing from the watch
	go func() {
		// Wait for opened signal
		<-opened

		// Show loading pop up
		loading := newLoadingPopUp(w)
		loading.Show()

		// Read root directory
		ls, err := client.ReadDir(ctx, "/")
		if err != nil {
			guiErr(err, "Error reading directory", false, w)
			return
		}
		// Set ls binding
		lsData.Set(lsToAny(ls))

		// Hide loading pop up
		loading.Hide()
	}()

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(
			theme.ViewRefreshIcon(),
			func() {
				refresh(ctx, cwdData, lsData, client, w, c)
			},
		),
		widget.NewToolbarAction(
			theme.FileApplicationIcon(),
			func() {
				dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
					if err != nil || uc == nil {
						return
					}

					resPath := uc.URI().Path()
					uc.Close()

					progressDlg := newProgress(w)
					progressDlg.Show()

					progCh, err := client.LoadResources(ctx, resPath)
					if err != nil {
						guiErr(err, "Error loading resources", false, w)
						return
					}

					for evt := range progCh {
						if evt.Err != nil {
							guiErr(evt.Err, "Error loading resources", false, w)
							return
						}

						switch evt.Operation {
						case infinitime.ResourceOperationRemoveObsolete:
							progressDlg.SetText("Removing " + evt.Name)
						case infinitime.ResourceOperationUpload:
							progressDlg.SetText("Uploading " + evt.Name)
							progressDlg.SetTotal(float64(evt.Total))
							progressDlg.SetValue(float64(evt.Sent))
						}
					}

					progressDlg.Hide()
					refresh(ctx, cwdData, lsData, client, w, c)
				}, w)
				dlg.SetConfirmText("Upload Resources")
				dlg.SetFilter(storage.NewExtensionFileFilter([]string{
					".zip",
				}))
				dlg.Show()
			},
		),
		widget.NewToolbarAction(
			theme.UploadIcon(),
			func() {
				// Create open dialog for file that will be uploaded
				dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
					if err != nil || uc == nil {
						return
					}
					// Get filepath and close
					localPath := uc.URI().Path()
					uc.Close()

					// Create new entry to store filepath
					filenameEntry := widget.NewEntry()
					// Set entry text to the file name of the selected file
					filenameEntry.SetText(filepath.Base(localPath))
					// Create new dialog asking for the filename of the file to be stored on the watch
					uploadDlg := dialog.NewForm("Upload", "Upload", "Cancel", []*widget.FormItem{
						widget.NewFormItem("Filename", filenameEntry),
					}, func(ok bool) {
						if !ok {
							return
						}

						// Get current directory
						cwd, _ := cwdData.Get()
						// Get remote path by joining current directory with filename
						remotePath := filepath.Join(cwd, filenameEntry.Text)

						// Create new progress dialog
						progressDlg := newProgress(w)
						progressDlg.Show()

						// Upload file
						progressCh, err := client.Upload(ctx, remotePath, localPath)
						if err != nil {
							guiErr(err, "Error uploading file", false, w)
							return
						}

						for progressEvt := range progressCh {
							progressDlg.SetTotal(float64(progressEvt.Total))
							progressDlg.SetValue(float64(progressEvt.Sent))
							if progressEvt.Sent == progressEvt.Total {
								break
							}
						}

						// Close progress dialog
						progressDlg.Hide()

						// Add file to listing (avoids full refresh)
						lsData.Append(api.FileInfo{
							IsDir: false,
							Name:  filepath.Base(remotePath),
						})
					}, w)
					uploadDlg.Show()
				}, w)
				dlg.Show()
			},
		),
		widget.NewToolbarAction(
			theme.FolderNewIcon(),
			func() {
				// Create new entry for filename
				filenameEntry := widget.NewEntry()
				// Create new dialog to ask for the filename
				mkdirDialog := dialog.NewForm("Make Directory", "Create", "Cancel", []*widget.FormItem{
					widget.NewFormItem("Filename", filenameEntry),
				}, func(ok bool) {
					if !ok {
						return
					}

					// Get current directory
					cwd, _ := cwdData.Get()
					// Get remote path by joining current directory and filename
					remotePath := filepath.Join(cwd, filenameEntry.Text)

					// Make directory
					err := client.Mkdir(ctx, remotePath)
					if err != nil {
						guiErr(err, "Error creating directory", false, w)
						return
					}

					// Add directory to listing (avoids full refresh)
					lsData.Append(api.FileInfo{
						IsDir: true,
						Name:  filepath.Base(remotePath),
					})
				}, w)
				mkdirDialog.Show()
			},
		),
	)

	// Add listener to listing data to create the new items on the GUI
	// whenever the listing changes
	lsData.AddListener(binding.NewDataListener(func() {
		c.Objects = makeItems(ctx, client, lsData, cwdData, w, c)
		c.Refresh()
	}))

	return container.NewBorder(
		nil,
		toolbar,
		nil,
		nil,
		container.NewVScroll(c),
	)
}

// makeItems creates GUI objects from listing data
func makeItems(
	ctx context.Context,
	client *api.Client,
	lsData binding.UntypedList,
	cwdData binding.String,
	w fyne.Window,
	c *fyne.Container,
) []fyne.CanvasObject {
	// Get listing data
	ls, _ := lsData.Get()

	// Create output slice with dame length as listing
	out := make([]fyne.CanvasObject, len(ls))
	for index, val := range ls {
		// Assert value as file info
		item := val.(api.FileInfo)

		var icon fyne.Resource
		// Decide which icon to use
		if item.IsDir {
			if item.Name == ".." {
				icon = theme.NavigateBackIcon()
			} else {
				icon = theme.FolderIcon()
			}
		} else {
			icon = theme.FileIcon()
		}

		// Create new button with the decided icon and the item name
		btn := widget.NewButtonWithIcon(item.Name, icon, nil)
		// Align left
		btn.Alignment = widget.ButtonAlignLeading
		// Decide which callback function to use
		if item.IsDir {
			btn.OnTapped = func() {
				// Get current directory
				cwd, _ := cwdData.Get()
				// Join current directory with item name
				cwd = filepath.Join(cwd, item.Name)
				// Set new current directory
				cwdData.Set(cwd)
				// Refresh GUI to display new directory
				refresh(ctx, cwdData, lsData, client, w, c)
			}
		} else {
			btn.OnTapped = func() {
				// Get current directory
				cwd, _ := cwdData.Get()
				// Join current directory with item name
				remotePath := filepath.Join(cwd, item.Name)
				// Create new save dialog
				dlg := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
					if err != nil || uc == nil {
						return
					}
					// Get path of selected file
					localPath := uc.URI().Path()
					// Close WriteCloser (it's not needed)
					uc.Close()

					// Create new progress dialog
					progressDlg := newProgress(w)
					progressDlg.Show()

					// Download file
					progressCh, err := client.Download(ctx, localPath, remotePath)
					if err != nil {
						guiErr(err, "Error downloading file", false, w)
						return
					}

					// For every progress event
					for progressEvt := range progressCh {
						progressDlg.SetTotal(float64(progressEvt.Total))
						progressDlg.SetValue(float64(progressEvt.Sent))
					}

					// Close progress dialog
					progressDlg.Hide()
				}, w)
				// Set filename to the item name
				dlg.SetFileName(item.Name)
				dlg.Show()
			}
		}

		if item.Name == ".." {
			out[index] = btn
			continue
		}

		moveBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
			moveEntry := widget.NewEntry()
			dlg := dialog.NewForm("Move", "Move", "Cancel", []*widget.FormItem{
				widget.NewFormItem("New Path", moveEntry),
			}, func(ok bool) {
				if !ok {
					return
				}

				// Get current directory
				cwd, _ := cwdData.Get()
				// Join current directory with item name
				oldPath := filepath.Join(cwd, item.Name)

				// Rename file
				err := client.Rename(ctx, oldPath, moveEntry.Text)
				if err != nil {
					guiErr(err, "Error renaming file", false, w)
					return
				}

				// Refresh GUI
				refresh(ctx, cwdData, lsData, client, w, c)
			}, w)
			dlg.Show()
		})

		removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			// Get current directory
			cwd, _ := cwdData.Get()
			// Join current directory with item name
			path := filepath.Join(cwd, item.Name)

			// Remove file
			err := client.Remove(ctx, path)
			if err != nil {
				guiErr(err, "Error removing file", false, w)
				return
			}

			// Refresh GUI
			refresh(ctx, cwdData, lsData, client, w, c)
		})

		// Add button to GUI component list
		out[index] = container.NewBorder(
			nil,
			nil,
			nil,
			container.NewHBox(moveBtn, removeBtn),
			btn,
		)
	}
	return out
}

func refresh(
	ctx context.Context,
	cwdData binding.String,
	lsData binding.UntypedList,
	client *api.Client,
	w fyne.Window,
	c *fyne.Container,
) {
	// Create and show new loading pop up
	loading := newLoadingPopUp(w)
	loading.Show()
	// Close pop up at the end of the function
	defer loading.Hide()

	// Get current directory
	cwd, _ := cwdData.Get()
	// Read directory
	ls, err := client.ReadDir(ctx, cwd)
	if err != nil {
		guiErr(err, "Error reading directory", false, w)
		return
	}
	// Set new listing data
	lsData.Set(lsToAny(ls))
	// Create new GUI objects
	c.Objects = makeItems(ctx, client, lsData, cwdData, w, c)
	// Refresh GUI
	c.Refresh()
}

func lsToAny(ls []api.FileInfo) []interface{} {
	out := make([]interface{}, len(ls)-1)
	for i, e := range ls {
		// Skip first element as it is always "."
		if i == 0 {
			continue
		}
		out[i-1] = e
	}
	return out
}
