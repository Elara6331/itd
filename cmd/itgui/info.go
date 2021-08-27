package main

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"net"

	"encoding/json"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"go.arsenm.dev/itd/internal/types"
)

func infoTab(parent fyne.Window) *fyne.Container {
	infoLayout := container.NewVBox(
		// Add rectangle for a bit of padding
		canvas.NewRectangle(color.Transparent),
	)

	// Create label for heart rate
	heartRateLbl := newText("0 BPM", 24)
	// Creae container to store heart rate section
	heartRate := container.NewVBox(
		newText("Heart Rate", 12),
		heartRateLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(heartRate)

	// Watch for heart rate updates
	go watch(types.ReqTypeWatchHeartRate, func(data interface{}) {
		// Change text of heart rate label
		heartRateLbl.Text = fmt.Sprintf("%d BPM", int(data.(float64)))
		// Refresh label
		heartRateLbl.Refresh()
	}, parent)

	// Create label for battery level
	battLevelLbl := newText("0%", 24)
	// Create container to store battery level section
	battLevel := container.NewVBox(
		newText("Battery Level", 12),
		battLevelLbl,
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(battLevel)

	// Watch for changes in battery level
	go watch(types.ReqTypeWatchBattLevel, func(data interface{}) {
		battLevelLbl.Text = fmt.Sprintf("%d%%", int(data.(float64)))
		battLevelLbl.Refresh()
	}, parent)

	fwVerString, err := get(types.ReqTypeFwVersion)
	if err != nil {
		guiErr(err, "Error getting firmware string", true, parent)
	}

	fwVer := container.NewVBox(
		newText("Firmware Version", 12),
		newText(fwVerString.(string), 24),
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(fwVer)

	btAddrString, err := get(types.ReqTypeBtAddress)
	if err != nil {
		panic(err)
	}

	btAddr := container.NewVBox(
		newText("Bluetooth Address", 12),
		newText(btAddrString.(string), 24),
		canvas.NewLine(theme.ShadowColor()),
	)
	infoLayout.Add(btAddr)

	return infoLayout
}

func watch(req int, onRecv func(data interface{}), parent fyne.Window) error {
	conn, err := net.Dial("unix", SockPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	err = json.NewEncoder(conn).Encode(types.Request{
		Type: req,
	})
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		res, err := getResp(scanner.Bytes())
		if err != nil {
			guiErr(err, "Error getting response from connection", false, parent)
			continue
		}
		onRecv(res.Value)
	}
	return nil
}

func get(req int) (interface{}, error) {
	conn, err := net.Dial("unix", SockPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	err = json.NewEncoder(conn).Encode(types.Request{
		Type: req,
	})
	if err != nil {
		return nil, err
	}
	line, _, err := bufio.NewReader(conn).ReadLine()
	if err != nil {
		return nil, err
	}
	res, err := getResp(line)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func getResp(line []byte) (*types.Response, error) {
	var res types.Response
	err := json.Unmarshal(line, &res)
	if err != nil {
		return nil, err
	}
	if res.Error {
		return nil, errors.New(res.Message)
	}
	return &res, nil
}

func newText(t string, size float32) *canvas.Text {
	text := canvas.NewText(t, theme.ForegroundColor())
	text.TextSize = size
	return text
}
