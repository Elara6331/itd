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

	"go.elara.ws/infinitime"
	"go.elara.ws/itd/mpris"
	"go.elara.ws/itd/translit"
	"go.elara.ws/logger/log"
)

func initMusicCtrl(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	mpris.Init(ctx)

	maps := k.Strings("notifs.translit.use")
	translit.Transliterators["custom"] = translit.Map(k.Strings("notifs.translit.custom"))

	mpris.OnChange(func(ct mpris.ChangeType, val string) {
		newVal := translit.Transliterate(val, maps...)
		if !firmwareUpdating {
			switch ct {
			case mpris.ChangeTypeStatus:
				dev.Music.SetStatus(val == "Playing")
			case mpris.ChangeTypeTitle:
				dev.Music.SetTrack(newVal)
			case mpris.ChangeTypeAlbum:
				dev.Music.SetAlbum(newVal)
			case mpris.ChangeTypeArtist:
				dev.Music.SetArtist(newVal)
			}
		}
	})

	// Watch for music events
	musicEvtCh, err := dev.Music.WatchEvents()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done("musicCtrl")
		// For every music event received
		for {
			select {
			case musicEvt := <-musicEvtCh:
				// Perform appropriate action based on event
				switch musicEvt {
				case infinitime.MusicEventPlay:
					mpris.Play()
				case infinitime.MusicEventPause:
					mpris.Pause()
				case infinitime.MusicEventNext:
					mpris.Next()
				case infinitime.MusicEventPrev:
					mpris.Prev()
				case infinitime.MusicEventVolUp:
					mpris.VolUp(uint(k.Int("music.vol.interval")))
				case infinitime.MusicEventVolDown:
					mpris.VolDown(uint(k.Int("music.vol.interval")))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Log completed initialization
	log.Info("Initialized InfiniTime music controls").Send()

	return nil
}
