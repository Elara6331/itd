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
	"log/slog"

	"go.elara.ws/itd/infinitime"
	"go.elara.ws/itd/mpris"
	"go.elara.ws/itd/translit"
)

func initMusicCtrl(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	mpris.Init(ctx)

	maps := cfg.Notifs.Translit.Use
	translit.Transliterators["custom"] = translit.Map(cfg.Notifs.Translit.Custom)

	mpris.OnChange(func(ct mpris.ChangeType, val string) {
		newVal := translit.Transliterate(val, maps...)
		if !firmwareUpdating {
			switch ct {
			case mpris.ChangeTypeStatus:
				dev.SetMusicStatus(val == "Playing")
			case mpris.ChangeTypeTitle:
				dev.SetMusicTrack(newVal)
			case mpris.ChangeTypeAlbum:
				dev.SetMusicAlbum(newVal)
			case mpris.ChangeTypeArtist:
				dev.SetMusicArtist(newVal)
			}
		}
	})

	// Watch for music events
	err := dev.WatchMusicEvents(ctx, func(event infinitime.MusicEvent, err error) {
		if err != nil {
			log.Error("Music event error", slog.Any("error", err))
		}
		
		// Perform appropriate action based on event
		switch event {
		case infinitime.MusicEventPlay:
			mpris.Play()
		case infinitime.MusicEventPause:
			mpris.Pause()
		case infinitime.MusicEventNext:
			mpris.Next()
		case infinitime.MusicEventPrev:
			mpris.Prev()
		case infinitime.MusicEventVolUp:
			mpris.VolUp(cfg.Music.Vol.Interval)
		case infinitime.MusicEventVolDown:
			mpris.VolDown(cfg.Music.Vol.Interval)
		}
	})
	if err != nil {
		return err
	}

	// Log completed initialization
	log.Info("Initialized InfiniTime music controls")

	return nil
}
