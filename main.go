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
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.arsenm.dev/infinitime"
)

var firmwareUpdating = false

func init() {
	// Set up logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set config settings
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("$HOME/.config")
	viper.SetConfigName("itd")
	viper.SetConfigType("toml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("itd")
	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Could not read in config")
	}
	viper.AutomaticEnv()
}

func main() {
	// Cleanly exit after function
	defer infinitime.Exit()

	// Connect to InfiniTime with default options
	dev, err := infinitime.Connect(&infinitime.Options{
		AttemptReconnect: viper.GetBool("conn.reconnect"),
	})
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to InfiniTime")
	}

	// When InfiniTime reconnects
	dev.OnReconnect(func() {
		// Set time to current time
		err = dev.SetTime(time.Now())
		if err != nil {
			log.Error().Err(err).Msg("Error setting current time on connected InfiniTime")
		}

		// If config specifies to notify on reconnect
		if viper.GetBool("notify.onReconnect") {
			// Send notification to InfiniTime
			err = dev.Notify("itd", "Successfully reconnected")
			if err != nil {
				log.Error().Err(err).Msg("Error sending notification to InfiniTime")
			}
		}
	})

	// Get firmware version
	ver, err := dev.Version()
	if err != nil {
		log.Error().Err(err).Msg("Error getting firmware version")
	}

	// Log connection
	log.Info().Str("version", ver).Msg("Connected to InfiniTime")

	// If config specifies to notify on connect
	if viper.GetBool("notify.onConnect") {
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

	// Initialize notification relay
	err = initNotifRelay(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing notification relay")
	}

	// Start control socket
	err = startSocket(dev)
	if err != nil {
		log.Error().Err(err).Msg("Error starting socket")
	}

	// Block forever
	select {}
}
