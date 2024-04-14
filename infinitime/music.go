package infinitime

import "context"

type MusicEvent uint8

const (
	MusicEventOpen    MusicEvent = 0xe0
	MusicEventPlay    MusicEvent = 0x00
	MusicEventPause   MusicEvent = 0x01
	MusicEventNext    MusicEvent = 0x03
	MusicEventPrev    MusicEvent = 0x04
	MusicEventVolUp   MusicEvent = 0x05
	MusicEventVolDown MusicEvent = 0x06
)

// SetMusicStatus sets whether the music is playing or paused.
func (d *Device) SetMusicStatus(playing bool) error {
	char, err := d.getChar(musicStatusChar)
	if err != nil {
		return err
	}

	if playing {
		_, err = char.WriteWithoutResponse([]byte{0x1})
	} else {
		_, err = char.WriteWithoutResponse([]byte{0x0})
	}
	return err
}

// SetMusicArtist sets the music artist.
func (d *Device) SetMusicArtist(artist string) error {
	char, err := d.getChar(musicArtistChar)
	if err != nil {
		return err
	}

	_, err = char.WriteWithoutResponse([]byte(artist))
	return err
}

// SetMusicTrack sets the music track name.
func (d *Device) SetMusicTrack(track string) error {
	char, err := d.getChar(musicTrackChar)
	if err != nil {
		return err
	}

	_, err = char.WriteWithoutResponse([]byte(track))
	return err
}

// SetMusicAlbum sets the music album name.
func (d *Device) SetMusicAlbum(album string) error {
	char, err := d.getChar(musicAlbumChar)
	if err != nil {
		return err
	}

	_, err = char.WriteWithoutResponse([]byte(album))
	return err
}

// WatchMusicEvents calls fn whenever the InfiniTime music app broadcasts an event.
func (d *Device) WatchMusicEvents(ctx context.Context, fn func(event MusicEvent, err error)) error {
	return watchChar(ctx, d, musicEventChar, fn)
}
