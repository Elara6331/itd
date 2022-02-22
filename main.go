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
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/knadh/koanf"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
)

var k = koanf.New(".")

//go:embed version.txt
var version string

var (
	firmwareUpdating = false
	// The FS must be updated when the watch is reconnected
	updateFS = false
)

func main() {
	infinitime.Init()
	// Cleanly exit after function
	defer infinitime.Exit()

	// Create infinitime options struct
	opts := &infinitime.Options{
		AttemptReconnect: k.Bool("conn.reconnect"),
		WhitelistEnabled: k.Bool("conn.whitelist.enabled"),
		Whitelist:        k.Strings("conn.whitelist.devices"),
		OnReqPasskey:     onReqPasskey,
	}

	// Connect to InfiniTime with default options
	dev, err := infinitime.Connect(opts)
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to InfiniTime")
	}

	// When InfiniTime reconnects
	opts.OnReconnect = func() {
		if k.Bool("on.reconnect.setTime") {
			// Set time to current time
			err = dev.SetTime(time.Now())
			if err != nil {
				return
			}
		}

		// If config specifies to notify on reconnect
		if k.Bool("on.reconnect.notify") {
			// Send notification to InfiniTime
			err = dev.Notify("itd", "Successfully reconnected")
			if err != nil {
				return
			}
		}

		// FS must be updated on reconnect
		updateFS = true
		// Resend weather on reconnect
		sendWeatherCh <- struct{}{}
	}

	// Get firmware version
	ver, err := dev.Version()
	if err != nil {
		log.Error().Err(err).Msg("Error getting firmware version")
	}

	// Log connection
	log.Info().Str("version", ver).Msg("Connected to InfiniTime")

	// If config specifies to notify on connect
	if k.Bool("on.connect.notify") {
		// Send notification to InfiniTime
		err = dev.Notify("itd", "Successfully connected")
		if err != nil {
			log.Error().Err(err).Msg("Error sending notification to InfiniTime")
		}
	}

	// Set time to current time
	err = dev.SetTime(time.Now())
	if err != nil {
		log.Error().Err(err).Msg("Error setting current time on connected InfiniTime")
	}

	// Initialize music controls
	err = initMusicCtrl(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing music control")
	}

	// Start control socket
	err = initCallNotifs(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing call notifications")
	}

	// Initialize notification relay
	err = initNotifRelay(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing notification relay")
	}

	// Initializa weather
	err = initWeather(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing weather")
	}

	// Start control socket
	err = startSocket(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error starting socket")
	}

	// Block forever
	select {}
}

func onReqPasskey() (uint32, error) {
	var out uint32
	if isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Print("Passkey: ")
		_, err := fmt.Scanln(&out)
		if err != nil {
			return 0, err
		}
	} else {
		passkey, ok, err := dlgs.Entry("Pairing", "Enter the passkey displayed on your watch.", "")
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, nil
		}
		passkeyInt, err := strconv.Atoi(passkey)
		return uint32(passkeyInt), err
	}
	return out, nil
}
