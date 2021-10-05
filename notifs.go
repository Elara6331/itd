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
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/translit"
)

func initNotifRelay(dev *infinitime.Device) error {
	// Connect to dbus session bus
	bus, err := dbus.SessionBus()
	if err != nil {
		return err
	}

	// Define rules to listen for
	var rules = []string{
		"type='method_call',member='Notify',path='/org/freedesktop/Notifications',interface='org.freedesktop.Notifications'",
	}
	var flag uint = 0
	// Becode monitor for notifications
	call := bus.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, rules, flag)
	if call.Err != nil {
		return call.Err
	}

	// Create channel to store notifications
	notifCh := make(chan *dbus.Message, 10)
	// Send events to channel
	bus.Eavesdrop(notifCh)

	go func() {
		// For every event sent to channel
		for v := range notifCh {
			// If firmware is updating, skip
			if firmwareUpdating {
				continue
			}

			// If body does not contain 5 elements, skip
			if len(v.Body) < 5 {
				continue
			}

			// Get requred fields
			sender, summary, body := v.Body[0].(string), v.Body[3].(string), v.Body[4].(string)

			// If fields are ignored in config, skip
			if ignored(sender, summary, body) {
				continue
			}

			maps := viper.GetStringSlice("notifs.translit.use")
			translit.Transliterators["custom"] = translit.Map(viper.GetStringSlice("notifs.translit.custom"))
			sender = translit.Transliterate(sender, maps...)
			summary = translit.Transliterate(summary, maps...)
			body = translit.Transliterate(body, maps...)

			var msg string
			// If summary does not exist, set message to body.
			// If it does, set message to summary, two newlines, and then body
			if summary == "" {
				msg = body
			} else {
				msg = fmt.Sprintf("%s\n\n%s", summary, body)
			}

			dev.Notify(sender, msg)
		}
	}()

	log.Info().Msg("Relaying notifications to InfiniTime")
	return nil
}

// ignored checks whether any fields were ignored in the config
func ignored(sender, summary, body string) bool {
	ignoreSender := viper.GetStringSlice("notifs.ignore.sender")
	ignoreSummary := viper.GetStringSlice("notifs.ignore.summary")
	ignoreBody := viper.GetStringSlice("notifs.ignore.body")
	return strSlcContains(ignoreSender, sender) ||
		strSlcContains(ignoreSummary, summary) ||
		strSlcContains(ignoreBody, body)
}

// strSliceContains checks whether a string slice contains a string
func strSlcContains(ss []string, s string) bool {
	for _, str := range ss {
		if str == s {
			return true
		}
	}
	return false
}
