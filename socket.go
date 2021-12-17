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
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/blefs"
	"go.arsenm.dev/itd/internal/types"
	"go.arsenm.dev/itd/translit"
)

type DoneMap map[string]chan struct{}

func (dm DoneMap) Exists(key string) bool {
	_, ok := dm[key]
	return ok
}

func (dm DoneMap) Done(key string) {
	ch := dm[key]
	ch <- struct{}{}
}

func (dm DoneMap) Create(key string) {
	dm[key] = make(chan struct{}, 1)
}

func (dm DoneMap) Remove(key string) {
	close(dm[key])
	delete(dm, key)
}

var done = DoneMap{}

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

	fs, err := dev.FS()
	if err != nil {
		log.Warn().Err(err).Msg("Error getting BLE filesystem")
	}

	go func() {
		for {
			// Accept socket connection
			conn, err := ln.Accept()
			if err != nil {
				log.Error().Err(err).Msg("Error accepting connection")
			}

			// Concurrently handle connection
			go handleConnection(conn, dev, fs)
		}
	}()

	// Log socket start
	log.Info().Str("path", viper.GetString("socket.path")).Msg("Started control socket")

	return nil
}

func handleConnection(conn net.Conn, dev *infinitime.Device, fs *blefs.FS) {
	defer conn.Close()

	// If an FS update is required (reconnect ocurred)
	if updateFS {
		// Get new FS
		newFS, err := dev.FS()
		if err != nil {
			fs = nil
			log.Warn().Err(err).Msg("Error updating BLE filesystem")
		}
		// Set FS pointer to new FS
		*fs = *newFS
		// Reset updateFS
		updateFS = false
	}

	// Create new scanner on connection
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req types.Request
		// Decode scanned message into types.Request
		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			connErr(conn, req.Type, err, "Error decoding JSON input")
			continue
		}

		// If firmware is updating, return error
		if firmwareUpdating {
			connErr(conn, req.Type, nil, "Firmware update in progress")
			return
		}

		switch req.Type {
		case types.ReqTypeHeartRate:
			// Get heart rate from watch
			heartRate, err := dev.HeartRate()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting heart rate")
				break
			}
			// Encode heart rate to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  req.Type,
				Value: heartRate,
			})
		case types.ReqTypeWatchHeartRate:
			heartRateCh, cancel, err := dev.WatchHeartRate()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting heart rate channel")
				break
			}
			reqID := uuid.New().String()
			go func() {
				done.Create(reqID)
				// For every heart rate value
				for heartRate := range heartRateCh {
					select {
					case <-done[reqID]:
						// Stop notifications if done signal received
						cancel()
						done.Remove(reqID)
						return
					default:
						// Encode response to connection if no done signal received
						json.NewEncoder(conn).Encode(types.Response{
							Type:  req.Type,
							ID:    reqID,
							Value: heartRate,
						})
					}
				}
			}()
		case types.ReqTypeBattLevel:
			// Get battery level from watch
			battLevel, err := dev.BatteryLevel()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting battery level")
				break
			}
			// Encode battery level to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  req.Type,
				Value: battLevel,
			})
		case types.ReqTypeWatchBattLevel:
			battLevelCh, cancel, err := dev.WatchBatteryLevel()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting battery level channel")
				break
			}
			reqID := uuid.New().String()
			go func() {
				done.Create(reqID)
				// For every battery level value
				for battLevel := range battLevelCh {
					select {
					case <-done[reqID]:
						// Stop notifications if done signal received
						cancel()
						done.Remove(reqID)
						return
					default:
						// Encode response to connection if no done signal received
						json.NewEncoder(conn).Encode(types.Response{
							Type:  req.Type,
							ID:    reqID,
							Value: battLevel,
						})
					}
				}
			}()
		case types.ReqTypeMotion:
			// Get battery level from watch
			motionVals, err := dev.Motion()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting motion values")
				break
			}
			// Encode battery level to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  req.Type,
				Value: motionVals,
			})
		case types.ReqTypeWatchMotion:
			motionValCh, cancel, err := dev.WatchMotion()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting heart rate channel")
				break
			}
			reqID := uuid.New().String()
			go func() {
				done.Create(reqID)
				// For every motion event
				for motionVals := range motionValCh {
					select {
					case <-done[reqID]:
						// Stop notifications if done signal received
						cancel()
						done.Remove(reqID)

						return
					default:
						// Encode response to connection if no done signal received
						json.NewEncoder(conn).Encode(types.Response{
							Type:  req.Type,
							ID:    reqID,
							Value: motionVals,
						})
					}
				}
			}()
		case types.ReqTypeStepCount:
			// Get battery level from watch
			stepCount, err := dev.StepCount()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting step count")
				break
			}
			// Encode battery level to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  req.Type,
				Value: stepCount,
			})
		case types.ReqTypeWatchStepCount:
			stepCountCh, cancel, err := dev.WatchStepCount()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting heart rate channel")
				break
			}
			reqID := uuid.New().String()
			go func() {
				done.Create(reqID)
				// For every step count value
				for stepCount := range stepCountCh {
					select {
					case <-done[reqID]:
						// Stop notifications if done signal received
						cancel()
						done.Remove(reqID)
						return
					default:
						// Encode response to connection if no done signal received
						json.NewEncoder(conn).Encode(types.Response{
							Type:  req.Type,
							ID:    reqID,
							Value: stepCount,
						})
					}
				}
			}()
		case types.ReqTypeFwVersion:
			// Get firmware version from watch
			version, err := dev.Version()
			if err != nil {
				connErr(conn, req.Type, err, "Error getting firmware version")
				break
			}
			// Encode version to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  req.Type,
				Value: version,
			})
		case types.ReqTypeBtAddress:
			// Encode bluetooth address to connection
			json.NewEncoder(conn).Encode(types.Response{
				Type:  req.Type,
				Value: dev.Address(),
			})
		case types.ReqTypeNotify:
			// If no data, return error
			if req.Data == nil {
				connErr(conn, req.Type, nil, "Data required for notify request")
				break
			}
			var reqData types.ReqDataNotify
			// Decode data map to notify request data
			err = mapstructure.Decode(req.Data, &reqData)
			if err != nil {
				connErr(conn, req.Type, err, "Error decoding request data")
				break
			}
			maps := viper.GetStringSlice("notifs.translit.use")
			translit.Transliterators["custom"] = translit.Map(viper.GetStringSlice("notifs.translit.custom"))
			title := translit.Transliterate(reqData.Title, maps...)
			body := translit.Transliterate(reqData.Body, maps...)
			// Send notification to watch
			err = dev.Notify(title, body)
			if err != nil {
				connErr(conn, req.Type, err, "Error sending notification")
				break
			}
			// Encode empty types.Response to connection
			json.NewEncoder(conn).Encode(types.Response{Type: req.Type})
		case types.ReqTypeSetTime:
			// If no data, return error
			if req.Data == nil {
				connErr(conn, req.Type, nil, "Data required for settime request")
				break
			}
			// Get string from data or return error
			reqTimeStr, ok := req.Data.(string)
			if !ok {
				connErr(conn, req.Type, nil, "Data for settime request must be RFC3339 formatted time string")
				break
			}

			var reqTime time.Time
			if reqTimeStr == "now" {
				reqTime = time.Now()
			} else {
				// Parse time as RFC3339/ISO8601
				reqTime, err = time.Parse(time.RFC3339, reqTimeStr)
				if err != nil {
					connErr(conn, req.Type, err, "Invalid time format. Time string must be formatted as ISO8601 or the word `now`")
					break
				}
			}
			// Set time on watch
			err = dev.SetTime(reqTime)
			if err != nil {
				connErr(conn, req.Type, err, "Error setting device time")
				break
			}
			// Encode empty types.Response to connection
			json.NewEncoder(conn).Encode(types.Response{Type: req.Type})
		case types.ReqTypeFwUpgrade:
			// If no data, return error
			if req.Data == nil {
				connErr(conn, req.Type, nil, "Data required for firmware upgrade request")
				break
			}
			var reqData types.ReqDataFwUpgrade
			// Decode data map to firmware upgrade request data
			err = mapstructure.Decode(req.Data, &reqData)
			if err != nil {
				connErr(conn, req.Type, err, "Error decoding request data")
				break
			}
			// Reset DFU to prepare for next update
			dev.DFU.Reset()
			switch reqData.Type {
			case types.UpgradeTypeArchive:
				// If less than one file, return error
				if len(reqData.Files) < 1 {
					connErr(conn, req.Type, nil, "Archive upgrade requires one file with .zip extension")
					break
				}
				// If file is not zip archive, return error
				if filepath.Ext(reqData.Files[0]) != ".zip" {
					connErr(conn, req.Type, nil, "Archive upgrade file must be a zip archive")
					break
				}
				// Load DFU archive
				err := dev.DFU.LoadArchive(reqData.Files[0])
				if err != nil {
					connErr(conn, req.Type, err, "Error loading archive file")
					break
				}
			case types.UpgradeTypeFiles:
				// If less than two files, return error
				if len(reqData.Files) < 2 {
					connErr(conn, req.Type, nil, "Files upgrade requires two files. First with .dat and second with .bin extension.")
					break
				}
				// If first file is not init packet, return error
				if filepath.Ext(reqData.Files[0]) != ".dat" {
					connErr(conn, req.Type, nil, "First file must be a .dat file")
					break
				}
				// If second file is not firmware image, return error
				if filepath.Ext(reqData.Files[1]) != ".bin" {
					connErr(conn, req.Type, nil, "Second file must be a .bin file")
					break
				}
				// Load individual DFU files
				err := dev.DFU.LoadFiles(reqData.Files[0], reqData.Files[1])
				if err != nil {
					connErr(conn, req.Type, err, "Error loading firmware files")
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
						Type:  req.Type,
						Value: event,
					})
				}
				firmwareUpdating = false
			}()

			// Set firmwareUpdating
			firmwareUpdating = true
			// Start DFU
			err = dev.DFU.Start()
			if err != nil {
				connErr(conn, req.Type, err, "Error performing upgrade")
				firmwareUpdating = false
				break
			}
			firmwareUpdating = false
		case types.ReqTypeFS:
			if fs == nil {
				connErr(conn, req.Type, nil, "BLE filesystem is not available")
				break
			}

			// If no data, return error
			if req.Data == nil {
				connErr(conn, req.Type, nil, "Data required for filesystem operations")
				break
			}

			var reqData types.ReqDataFS
			// Decode data map to firmware upgrade request data
			err = mapstructure.Decode(req.Data, &reqData)
			if err != nil {
				connErr(conn, req.Type, err, "Error decoding request data")
				break
			}

			// Clean input filepaths
			reqData.Files = cleanPaths(reqData.Files)

			switch reqData.Type {
			case types.FSTypeDelete:
				if len(reqData.Files) == 0 {
					connErr(conn, req.Type, nil, "Remove FS command requires at least one file")
					break
				}
				for _, file := range reqData.Files {
					err := fs.Remove(file)
					if err != nil {
						connErr(conn, req.Type, err, "Error removing file")
						break
					}
				}
				json.NewEncoder(conn).Encode(types.Response{Type: req.Type})
			case types.FSTypeMove:
				if len(reqData.Files) != 2 {
					connErr(conn, req.Type, nil, "Move FS command requires an old path and new path in the files list")
					break
				}
				err := fs.Rename(reqData.Files[0], reqData.Files[1])
				if err != nil {
					connErr(conn, req.Type, err, "Error moving file")
					break
				}
				json.NewEncoder(conn).Encode(types.Response{Type: req.Type})
			case types.FSTypeMkdir:
				if len(reqData.Files) == 0 {
					connErr(conn, req.Type, nil, "Mkdir FS command requires at least one file")
					break
				}
				for _, file := range reqData.Files {
					err := fs.Mkdir(file)
					if err != nil {
						connErr(conn, req.Type, err, "Error creating directory")
						break
					}
				}
				json.NewEncoder(conn).Encode(types.Response{Type: req.Type})
			case types.FSTypeList:
				if len(reqData.Files) != 1 {
					connErr(conn, req.Type, nil, "List FS command requires a path to list in the files list")
					break
				}
				entries, err := fs.ReadDir(reqData.Files[0])
				if err != nil {
					connErr(conn, req.Type, err, "Error reading directory")
					break
				}
				var out []types.FileInfo
				for _, entry := range entries {
					info, err := entry.Info()
					if err != nil {
						connErr(conn, req.Type, err, "Error getting file info")
						break
					}
					out = append(out, types.FileInfo{
						Name:  info.Name(),
						Size:  info.Size(),
						IsDir: info.IsDir(),
					})
				}
				json.NewEncoder(conn).Encode(types.Response{
					Type:  req.Type,
					Value: out,
				})
			case types.FSTypeWrite:
				if len(reqData.Files) != 2 {
					connErr(conn, req.Type, nil, "Write FS command requires a path to the file to write")
					break
				}

				localFile, err := os.Open(reqData.Files[1])
				if err != nil {
					connErr(conn, req.Type, err, "Error opening local file")
					break
				}
				defer localFile.Close()

				localInfo, err := localFile.Stat()
				if err != nil {
					connErr(conn, req.Type, err, "Error getting local file information")
					break
				}

				remoteFile, err := fs.Create(reqData.Files[0], uint32(localInfo.Size()))
				if err != nil {
					connErr(conn, req.Type, err, "Error creating remote file")
					break
				}
				defer remoteFile.Close()

				go func() {
					// For every progress event
					for sent := range remoteFile.Progress() {
						// Encode event on connection
						json.NewEncoder(conn).Encode(types.Response{
							Type: req.Type,
							Value: types.FSTransferProgress{
								Type:  types.FSTypeWrite,
								Total: remoteFile.Size(),
								Sent:  sent,
							},
						})
					}
				}()

				json.NewEncoder(conn).Encode(types.Response{Type: req.Type})

				io.Copy(remoteFile, localFile)

				json.NewEncoder(conn).Encode(types.Response{
					Type: req.Type,
					Value: types.FSTransferProgress{
						Type:  types.FSTypeWrite,
						Total: remoteFile.Size(),
						Sent:  remoteFile.Size(),
						Done:  true,
					},
				})
			case types.FSTypeRead:
				if len(reqData.Files) != 2 {
					connErr(conn, req.Type, nil, "Read FS command requires a path to the file to read")
					break
				}
				localFile, err := os.Create(reqData.Files[0])
				if err != nil {
					connErr(conn, req.Type, err, "Error creating local file")
					break
				}
				defer localFile.Close()

				remoteFile, err := fs.Open(reqData.Files[1])
				if err != nil {
					connErr(conn, req.Type, err, "Error opening remote file")
					break
				}
				defer remoteFile.Close()

				go func() {
					// For every progress event
					for rcvd := range remoteFile.Progress() {
						// Encode event on connection
						json.NewEncoder(conn).Encode(types.Response{
							Type: req.Type,
							Value: types.FSTransferProgress{
								Type:  types.FSTypeRead,
								Total: remoteFile.Size(),
								Sent:  rcvd,
							},
						})
					}
				}()

				json.NewEncoder(conn).Encode(types.Response{Type: req.Type})

				io.Copy(localFile, remoteFile)

				json.NewEncoder(conn).Encode(types.Response{
					Type: req.Type,
					Value: types.FSTransferProgress{
						Type:  types.FSTypeRead,
						Total: remoteFile.Size(),
						Sent:  remoteFile.Size(),
						Done:  true,
					},
				})
			}
		case types.ReqTypeCancel:
			if req.Data == nil {
				connErr(conn, req.Type, nil, "No data provided. Cancel request requires request ID string as data.")
				continue
			}
			reqID, ok := req.Data.(string)
			if !ok {
				connErr(conn, req.Type, nil, "Invalid data. Cancel request required request ID string as data.")
			}
			// Stop notifications
			done.Done(reqID)
			json.NewEncoder(conn).Encode(types.Response{Type: req.Type})
		default:
			connErr(conn, req.Type, nil, fmt.Sprintf("Unknown request type %d", req.Type))
		}
	}
}

func connErr(conn net.Conn, resType int, err error, msg string) {
	var res types.Response
	// If error exists, add to types.Response, otherwise don't
	if err != nil {
		log.Error().Err(err).Msg(msg)
		res = types.Response{Message: fmt.Sprintf("%s: %s", msg, err)}
	} else {
		log.Error().Msg(msg)
		res = types.Response{Message: msg, Type: resType}
	}
	res.Error = true

	// Encode error to connection
	json.NewEncoder(conn).Encode(res)
}

// cleanPaths runs strings.TrimSpace and filepath.Clean
// on all inputs, and returns the updated slice
func cleanPaths(paths []string) []string {
	for index, path := range paths {
		newPath := strings.TrimSpace(path)
		paths[index] = filepath.Clean(newPath)
	}
	return paths
}
