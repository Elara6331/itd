package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
)

func initCallNotifs(dev *infinitime.Device) error {
	// Define rule to filter dbus messages
	rule := "type='signal',sender='org.freedesktop.ModemManager1',interface='org.freedesktop.ModemManager1.Modem.Voice',member='CallAdded'"

	// Use dbus-monitor command with profiling output as a workaround
	// because go-bluetooth seems to monopolize the system bus connection
	// which makes monitoring show only bluez-related messages.
	cmd := exec.Command("dbus-monitor", "--system", "--profile", rule)
	// Get command output pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	// Run command asynchronously
	err = cmd.Start()
	if err != nil {
		return err
	}

	// Create new scanner for command output
	scanner := bufio.NewScanner(stdout)
	go func() {
		// For each line in output
		for scanner.Scan() {
			// Get line as string
			text := scanner.Text()

			// If line starts with "#", it is part of
			// the field format, skip it.
			if strings.HasPrefix(text, "#") {
				continue
			}

			// Split line into fields. The order is as follows:
			// type timestamp serial sender destination path interface member
			fields := strings.Fields(text)
			// Field 7 is Member. Make sure it is "CallAdded".
			if fields[7] == "CallAdded" {
				// Get Modem ID from modem path
				modemID := parseModemID(fields[5])
				// Get call ID of current call
				callID, err := getCurrentCallID(modemID)
				if err != nil {
					continue
				}
				// Get phone number of current call
				phoneNum, err := getPhoneNum(callID)
				if err != nil {
					continue
				}
				// Send call notification to PineTime
				resCh, err := dev.NotifyCall(phoneNum)
				if err != nil {
					continue
				}
				go func() {
					// Wait for PineTime response
					res := <-resCh
					switch res {
					case infinitime.CallStatusAccepted:
						// Attempt to accept call
						err = acceptCall(callID)
						if err != nil {
							log.Warn().Err(err).Msg("Error accepting call")
						}
					case infinitime.CallStatusDeclined:
						// Attempt to decline call
						err = declineCall(callID)
						if err != nil {
							log.Warn().Err(err).Msg("Error declining call")
						}
					case infinitime.CallStatusMuted:
						// Warn about unimplemented muting
						log.Warn().Msg("Muting calls is not implemented")
					}
				}()
			}
		}
	}()
	return nil
}

func parseModemID(modemPath string) int {
	// Split path by "/"
	splitPath := strings.Split(modemPath, "/")
	// Get last element and convert to integer
	id, _ := strconv.Atoi(splitPath[len(splitPath)-1])
	return id
}

func getCurrentCallID(modemID int) (int, error) {
	// Create mmcli command
	cmd := exec.Command("mmcli", "--voice-list-calls", "-m", fmt.Sprint(modemID), "-J")
	// Run command and get output
	data, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var calls map[string][]string
	// Decode JSON from command output
	err = json.Unmarshal(data, &calls)
	if err != nil {
		return 0, err
	}
	// Get first call in output
	firstCall := calls["modem.voice.call"][0]
	// Split path by "/"
	splitCall := strings.Split(firstCall, "/")
	// Return last element converted to integer
	return strconv.Atoi(splitCall[len(splitCall)-1])
}

func getPhoneNum(callID int) (string, error) {
	// Create dbus-send command
	cmd := exec.Command("dbus-send",
		"--dest=org.freedesktop.ModemManager1",
		"--system",
		"--print-reply=literal",
		"--type=method_call",
		fmt.Sprintf("/org/freedesktop/ModemManager1/Call/%d", callID),
		"org.freedesktop.DBus.Properties.Get",
		"string:org.freedesktop.ModemManager1.Call",
		"string:Number",
	)
	// Run command and get output
	numData, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Split output into fields
	num := strings.Fields(string(numData))
	// Return last field
	return num[len(num)-1], nil
}

func acceptCall(callID int) error {
	// Create dbus-send command
	cmd := exec.Command("dbus-send",
		"--dest=org.freedesktop.ModemManager1",
		"--print-reply",
		"--system",
		"--type=method_call",
		fmt.Sprintf("/org/freedesktop/ModemManager1/Call/%d", callID),
		"org.freedesktop.ModemManager1.Call.Accept",
	)
	// Run command and return errpr
	return cmd.Run()
}

func declineCall(callID int) error {
	// Create dbus-send command
	cmd := exec.Command("dbus-send",
		"--dest=org.freedesktop.ModemManager1",
		"--print-reply",
		"--system",
		"--type=method_call",
		fmt.Sprintf("/org/freedesktop/ModemManager1/Call/%d", callID),
		"org.freedesktop.ModemManager1.Call.Hangup",
	)
	// Run command and return errpr
	return cmd.Run()
}
