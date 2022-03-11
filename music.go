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
	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/pkg/player"
	"go.arsenm.dev/itd/translit"
)

func initMusicCtrl(dev *infinitime.Device) error {
	player.Init()

	maps := k.Strings("notifs.translit.use")
	translit.Transliterators["custom"] = translit.Map(k.Strings("notifs.translit.custom"))

	player.OnChange(func(ct player.ChangeType, val string) {
		newVal := translit.Transliterate(val, maps...)
		if !firmwareUpdating {
			switch ct {
			case player.ChangeTypeStatus:
				dev.Music.SetStatus(val == "Playing")
			case player.ChangeTypeTitle:
				dev.Music.SetTrack(newVal)
			case player.ChangeTypeAlbum:
				dev.Music.SetAlbum(newVal)
			case player.ChangeTypeArtist:
				dev.Music.SetArtist(newVal)
			}
		}
	})

	// Watch for music events
	musicEvtCh, err := dev.Music.WatchEvents()
	if err != nil {
		return err
	}
	go func() {
		// For every music event received
		for musicEvt := range musicEvtCh {
			// Perform appropriate action based on event
			switch musicEvt {
			case infinitime.MusicEventPlay:
				player.Play()
			case infinitime.MusicEventPause:
				player.Pause()
			case infinitime.MusicEventNext:
				player.Next()
			case infinitime.MusicEventPrev:
				player.Prev()
			case infinitime.MusicEventVolUp:
				player.VolUp(uint(k.Int("music.vol.interval")))
			case infinitime.MusicEventVolDown:
				player.VolDown(uint(k.Int("music.vol.interval")))
			}
		}
	}()

	// Log completed initialization
	log.Info().Msg("Initialized InfiniTime music controls")

	return nil
}
