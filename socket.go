/*
 *	itd uses bluetooth low energy to communicate with InfiniTime devices
 *	Copyright (C) 2021 Arsen Musayelyan
 *
 *	This program is free software: you can redistribute it and/or modify
 *	it under the terms of the GNU General Public License as published by
 *	the Free Software Foundation, either version 3 of the License, or
 *	(at your option) any later version.
 *
 *	This program is distributed in the hope that it will be useful,
 *	but WITHOUT ANY WARRANTY; without even the implied warranty of
 *	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *	GNU General Public License for more details.
 *
 *	You should have received a copy of the GNU General Public License
 *	along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/types"
	"go.arsenm.dev/itd/translit"
)

func startSocket(dev *infinitime.Device) error {
	// Make socket directory if non-existant
	err := os.MkdirAll(filepath.Dir(viper.GetString("socket.path")), 0755)
	if err != nil {
		return err
	}

	// Remove old socket if it exists
	err = os.RemoveAll(viper.GetString("socket.path"))
	if err != nil {
		return err
	}

	// Listen on socket path
	ln, err := net.Listen("unix", viper.GetString("socket.path"))
	if err != nil {
		return err
	}

	go func() {
		for {
			// Accept socket connection
			conn, err := ln.Accept()
			if err != nil {
				log.Error().Err(err).Msg("Error accepting connection")
			}

			// Concurrently handle connection
			go handleConnection(conn, dev)
		}
	}()

	// Log socket start
	log.Info().Str("path", viper.GetString("socket.path")).Msg("Started control socket")

	return nil
}

func handleConnection(conn net.Conn, dev *infinitime.Device) {
	defer conn.Close()
	// If firmware is updating, return error
	if firmwareUpdating {
		connErr(conn, nil, "Firmware update in progress")
		return
	}

	// Create new scanner on connection
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req types.Request
		// Decode scanned message into types.Request
		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			connErr(conn, err, "Error decoding JSON input")
			continue
		}

		switch req.Type {
		case types.ReqTypeHeartRate:
			// Get heart rate from watch
			heartRate, err := dev.HeartRate()
			if err != nil {
				connErr(conn, err, "Error getting heart rate")
				break
			}
			// Encode heart rate to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  types.ResTypeHeartRate,
				Value: heartRate,
			})
		case types.ReqTypeWatchHeartRate:
			heartRateCh, err := dev.WatchHeartRate()
			if err != nil {
				connErr(conn, err, "Error getting heart rate channel")
				break
			}
			go func() {
				for heartRate := range heartRateCh {
					json.NewEncoder(conn).Encode(types.Response{
						Type:  types.ResTypeWatchHeartRate,
						Value: heartRate,
					})
				}
			}()
		case types.ReqTypeBattLevel:
			// Get battery level from watch
			battLevel, err := dev.BatteryLevel()
			if err != nil {
				connErr(conn, err, "Error getting battery level")
				break
			}
			// Encode battery level to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  types.ResTypeBattLevel,
				Value: battLevel,
			})
		case types.ReqTypeWatchBattLevel:
			battLevelCh, err := dev.WatchBatteryLevel()
			if err != nil {
				connErr(conn, err, "Error getting battery level channel")
				break
			}
			go func() {
				for battLevel := range battLevelCh {
					json.NewEncoder(conn).Encode(types.Response{
						Type:  types.ResTypeWatchBattLevel,
						Value: battLevel,
					})
				}
			}()
		case types.ReqTypeMotion:
			// Get battery level from watch
			motionVals, err := dev.Motion()
			if err != nil {
				connErr(conn, err, "Error getting motion values")
				break
			}
			// Encode battery level to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  types.ResTypeMotion,
				Value: motionVals,
			})
		case types.ReqTypeWatchMotion:
			motionValCh, _, err := dev.WatchMotion()
			if err != nil {
				connErr(conn, err, "Error getting heart rate channel")
				break
			}
			go func() {
				for motionVals := range motionValCh {
					json.NewEncoder(conn).Encode(types.Response{
						Type:  types.ResTypeWatchMotion,
						Value: motionVals,
					})
				}
			}()
		case types.ReqTypeStepCount:
			// Get battery level from watch
			stepCount, err := dev.StepCount()
			if err != nil {
				connErr(conn, err, "Error getting step count")
				break
			}
			// Encode battery level to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  types.ResTypeStepCount,
				Value: stepCount,
			})
		case types.ReqTypeWatchStepCount:
			stepCountCh, _, err := dev.WatchStepCount()
			if err != nil {
				connErr(conn, err, "Error getting heart rate channel")
				break
			}
			go func() {
				for stepCount := range stepCountCh {
					json.NewEncoder(conn).Encode(types.Response{
						Type:  types.ResTypeWatchStepCount,
						Value: stepCount,
					})
				}
			}()
		case types.ReqTypeFwVersion:
			// Get firmware version from watch
			version, err := dev.Version()
			if err != nil {
				connErr(conn, err, "Error getting firmware version")
				break
			}
			// Encode version to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  types.ResTypeFwVersion,
				Value: version,
			})
		case types.ReqTypeBtAddress:
			// Encode bluetooth address to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  types.ResTypeBtAddress,
				Value: dev.Address(),
			})
		case types.ReqTypeNotify:
			// If no data, return error
			if req.Data == nil {
				connErr(conn, nil, "Data required for notify request")
				break
			}
			var reqData types.ReqDataNotify
			// Decode data map to notify request data
			err = mapstructure.Decode(req.Data, &reqData)
			if err != nil {
				connErr(conn, err, "Error decoding request data")
				break
			}
			maps := viper.GetStringSlice("notifs.translit.use")
			translit.Transliterators["custom"] = translit.Map(viper.GetStringSlice("notifs.translit.custom"))
			title := translit.Transliterate(reqData.Title, maps...)
			body := translit.Transliterate(reqData.Body, maps...)
			// Send notification to watch
			err = dev.Notify(title, body)
			if err != nil {
				connErr(conn, err, "Error sending notification")
				break
			}
			// Encode empty types.Response to connection
			json.NewEncoder(conn).Encode(types.Response{Type: types.ResTypeNotify})
		case types.ReqTypeSetTime:
			// If no data, return error
			if req.Data == nil {
				connErr(conn, nil, "Data required for settime request")
				break
			}
			// Get string from data or return error
			reqTimeStr, ok := req.Data.(string)
			if !ok {
				connErr(conn, nil, "Data for settime request must be RFC3339 formatted time string")
				break
			}

			var reqTime time.Time
			if reqTimeStr == "now" {
				reqTime = time.Now()
			} else {
				// Parse time as RFC3339/ISO8601
				reqTime, err = time.Parse(time.RFC3339, reqTimeStr)
				if err != nil {
					connErr(conn, err, "Invalid time format. Time string must be formatted as ISO8601 or the word `now`")
					break
				}
			}
			// Set time on watch
			err = dev.SetTime(reqTime)
			if err != nil {
				connErr(conn, err, "Error setting device time")
				break
			}
			// Encode empty types.Response to connection
			json.NewEncoder(conn).Encode(types.Response{Type: types.ResTypeSetTime})
		case types.ReqTypeFwUpgrade:
			// If no data, return error
			if req.Data == nil {
				connErr(conn, nil, "Data required for firmware upgrade request")
				break
			}
			var reqData types.ReqDataFwUpgrade
			// Decode data map to firmware upgrade request data
			err = mapstructure.Decode(req.Data, &reqData)
			if err != nil {
				connErr(conn, err, "Error decoding request data")
				break
			}
			// Reset DFU to prepare for next update
			dev.DFU.Reset()
			switch reqData.Type {
			case types.UpgradeTypeArchive:
				// If less than one file, return error
				if len(reqData.Files) < 1 {
					connErr(conn, nil, "Archive upgrade requires one file with .zip extension")
					break
				}
				// If file is not zip archive, return error
				if filepath.Ext(reqData.Files[0]) != ".zip" {
					connErr(conn, nil, "Archive upgrade file must be a zip archive")
					break
				}
				// Load DFU archive
				err := dev.DFU.LoadArchive(reqData.Files[0])
				if err != nil {
					connErr(conn, err, "Error loading archive file")
					break
				}
			case types.UpgradeTypeFiles:
				// If less than two files, return error
				if len(reqData.Files) < 2 {
					connErr(conn, nil, "Files upgrade requires two files. First with .dat and second with .bin extension.")
					break
				}
				// If first file is not init packet, return error
				if filepath.Ext(reqData.Files[0]) != ".dat" {
					connErr(conn, nil, "First file must be a .dat file")
					break
				}
				// If second file is not firmware image, return error
				if filepath.Ext(reqData.Files[1]) != ".bin" {
					connErr(conn, nil, "Second file must be a .bin file")
					break
				}
				// Load individual DFU files
				err := dev.DFU.LoadFiles(reqData.Files[0], reqData.Files[1])
				if err != nil {
					connErr(conn, err, "Error loading firmware files")
					break
				}
			}

			go func() {
				// Get progress
				progress := dev.DFU.Progress()
				// For every progress event
				for event := range progress {
					// Encode event on connection
					json.NewEncoder(conn).Encode(types.Response{
						Type:  types.ResTypeDFUProgress,
						Value: event,
					})
				}
			}()

			// Set firmwareUpdating
			firmwareUpdating = true
			// Start DFU
			err = dev.DFU.Start()
			if err != nil {
				connErr(conn, err, "Error performing upgrade")
				firmwareUpdating = false
				break
			}
			firmwareUpdating = false
		}
	}
}

func connErr(conn net.Conn, err error, msg string) {
	var res types.Response
	// If error exists, add to types.Response, otherwise don't
	if err != nil {
		log.Error().Err(err).Msg(msg)
		res = types.Response{Message: fmt.Sprintf("%s: %s", msg, err)}
	} else {
		log.Error().Msg(msg)
		res = types.Response{Message: msg}
	}
	res.Error = true

	// Encode error to connection
	json.NewEncoder(conn).Encode(res)
}
