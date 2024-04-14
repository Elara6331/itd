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
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"go.elara.ws/itd/infinitime"
	"go.elara.ws/itd/internal/utils"
	"go.elara.ws/itd/translit"
	"go.elara.ws/logger/log"
)

func initNotifRelay(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	// Connect to dbus session bus
	bus, err := utils.NewSessionBusConn(ctx)
	if err != nil {
		return err
	}

	// Define rules to listen for
	rules := []string{
		"type='method_call',member='Notify',path='/org/freedesktop/Notifications',interface='org.freedesktop.Notifications'",
	}
	var flag uint = 0
	// Becode monitor for notifications
	call := bus.BusObject().CallWithContext(
		ctx, "org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, rules, flag,
	)
	if call.Err != nil {
		return call.Err
	}

	// Create channel to store notifications
	notifCh := make(chan *dbus.Message, 10)
	// Send events to channel
	bus.Eavesdrop(notifCh)

	wg.Add(1)
	go func() {
		defer wg.Done("notifRelay")
		// For every event sent to channel
		for {
			select {
			case v := <-notifCh:
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

				maps := k.Strings("notifs.translit.use")
				translit.Transliterators["custom"] = translit.Map(k.Strings("notifs.translit.custom"))
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
			case <-ctx.Done():
				bus.Close()
				return
			}
		}
	}()

	log.Info("Relaying notifications to InfiniTime").Send()
	return nil
}

// ignored checks whether any fields were ignored in the config
func ignored(sender, summary, body string) bool {
	ignoreSender := k.Strings("notifs.ignore.sender")
	ignoreSummary := k.Strings("notifs.ignore.summary")
	ignoreBody := k.Strings("notifs.ignore.body")
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
