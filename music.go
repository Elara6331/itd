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
	"github.com/spf13/viper"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/pkg/player"
)

func initMusicCtrl(dev *infinitime.Device) error {
	// On player status change, set status
	err := player.Status(func(newStatus bool) {
		if !firmwareUpdating {
			dev.Music.SetStatus(newStatus)
		}
	})
	if err != nil {
		return err
	}

	// On player title change, set track
	err = player.Metadata("title", func(newTitle string) {
		if !firmwareUpdating {
			dev.Music.SetTrack(newTitle)
		}
	})
	if err != nil {
		return err
	}

	// On player album change, set album
	err = player.Metadata("album", func(newAlbum string) {
		if !firmwareUpdating {
			dev.Music.SetAlbum(newAlbum)
		}
	})
	if err != nil {
		return err
	}

	// On player artist change, set artist
	err = player.Metadata("artist", func(newArtist string) {
		if !firmwareUpdating {
			dev.Music.SetArtist(newArtist)
		}
	})
	if err != nil {
		return err
	}

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
				player.VolUp(viper.GetUint("music.vol.interval"))
			case infinitime.MusicEventVolDown:
				player.VolDown(viper.GetUint("music.vol.interval"))
			}
		}
	}()

	// Log completed initialization
	log.Info().Msg("Initialized InfiniTime music controls")

	return nil
}
