package mpris

import (
	"context"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"go.elara.ws/itd/internal/utils"
)

var (
	method, monitor *dbus.Conn
	monitorCh       chan *dbus.Message
	onChangeOnce    sync.Once
)

// Init makes required connections to DBus and
// initializes change monitoring channel
func Init(ctx context.Context) error {
	// Connect to session bus for monitoring
	monitorConn, err := utils.NewSessionBusConn(ctx)
	if err != nil {
		return err
	}
	// Add match rule for PropertiesChanged on media player
	monitorConn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/mpris/MediaPlayer2"),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	)
	monitorCh = make(chan *dbus.Message, 10)
	monitorConn.Eavesdrop(monitorCh)

	// Connect to session bus for method calls
	methodConn, err := utils.NewSessionBusConn(ctx)
	if err != nil {
		return err
	}
	method, monitor = methodConn, monitorConn
	return nil
}

// Exit closes all connections and channels
func Exit() {
	close(monitorCh)
	method.Close()
	monitor.Close()
}

// Play uses MPRIS to play media
func Play() error {
	player, err := getPlayerObj()
	if err != nil {
		return err
	}
	if player != nil {
		call := player.Call("org.mpris.MediaPlayer2.Player.Play", 0)
		if call.Err != nil {
			return call.Err
		}
	}
	return nil
}

// Pause uses MPRIS to pause media
func Pause() error {
	player, err := getPlayerObj()
	if err != nil {
		return err
	}
	if player != nil {
		call := player.Call("org.mpris.MediaPlayer2.Player.Pause", 0)
		if call.Err != nil {
			return call.Err
		}
	}
	return nil
}

// Next uses MPRIS to skip to next media
func Next() error {
	player, err := getPlayerObj()
	if err != nil {
		return err
	}
	if player != nil {
		call := player.Call("org.mpris.MediaPlayer2.Player.Next", 0)
		if call.Err != nil {
			return call.Err
		}
	}
	return nil
}

// Prev uses MPRIS to skip to previous media
func Prev() error {
	player, err := getPlayerObj()
	if err != nil {
		return err
	}
	if player != nil {
		call := player.Call("org.mpris.MediaPlayer2.Player.Previous", 0)
		if call.Err != nil {
			return call.Err
		}
	}
	return nil
}

func VolUp(percent uint) error {
	player, err := getPlayerObj()
	if err != nil {
		return err
	}
	if player != nil {
		currentVal, err := player.GetProperty("org.mpris.MediaPlayer2.Player.Volume")
		if err != nil {
			return err
		}
		newVal := currentVal.Value().(float64) + (float64(percent) / 100)
		err = player.SetProperty("org.mpris.MediaPlayer2.Player.Volume", newVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func VolDown(percent uint) error {
	player, err := getPlayerObj()
	if err != nil {
		return err
	}
	if player != nil {
		currentVal, err := player.GetProperty("org.mpris.MediaPlayer2.Player.Volume")
		if err != nil {
			return err
		}
		newVal := currentVal.Value().(float64) - (float64(percent) / 100)
		err = player.SetProperty("org.mpris.MediaPlayer2.Player.Volume", newVal)
		if err != nil {
			return err
		}
	}
	return nil
}

type ChangeType int

const (
	ChangeTypeTitle ChangeType = iota
	ChangeTypeArtist
	ChangeTypeAlbum
	ChangeTypeStatus
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeTitle:
		return "Title"
	case ChangeTypeAlbum:
		return "Album"
	case ChangeTypeArtist:
		return "Artist"
	case ChangeTypeStatus:
		return "Status"
	}
	return ""
}

// OnChange runs cb when a value changes
func OnChange(cb func(ChangeType, string)) {
	go onChangeOnce.Do(func() {
		// For every message on channel
		for msg := range monitorCh {
			// Parse PropertiesChanged
			iface, changed, ok := parsePropertiesChanged(msg)
			if !ok || iface != "org.mpris.MediaPlayer2.Player" {
				continue
			}

			// For every property changed
			for name, val := range changed {
				// If metadata changed
				if name == "Metadata" {
					// Get fields
					fields := val.Value().(map[string]dbus.Variant)
					// For every field
					for name, val := range fields {
						// Handle each field appropriately
						if strings.HasSuffix(name, "title") {
							title := val.Value().(string)
							if title == "" {
								title = "Unknown " + ChangeTypeTitle.String()
							}
							cb(ChangeTypeTitle, title)
						} else if strings.HasSuffix(name, "album") {
							album := val.Value().(string)
							if album == "" {
								album = "Unknown " + ChangeTypeAlbum.String()
							}
							cb(ChangeTypeAlbum, album)
						} else if strings.HasSuffix(name, "artist") {
							var artists string
							switch artistVal := val.Value().(type) {
							case string:
								artists = artistVal
							case []string:
								artists = strings.Join(artistVal, ", ")
							}
							if artists == "" {
								artists = "Unknown " + ChangeTypeArtist.String()
							}
							cb(ChangeTypeArtist, artists)
						}
					}
				} else if name == "PlaybackStatus" {
					// Handle status change
					cb(ChangeTypeStatus, val.Value().(string))
				}
			}
		}
	})
}

// getPlayerNames gets all DBus MPRIS player bus names
func getPlayerNames(conn *dbus.Conn) ([]string, error) {
	var names []string
	err := conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return nil, err
	}
	var players []string
	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2") {
			players = append(players, name)
		}
	}
	return players, nil
}

// GetPlayerObj gets the object corresponding to the first
// bus name found in DBus
func getPlayerObj() (dbus.BusObject, error) {
	players, err := getPlayerNames(method)
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		return nil, nil
	}
	return method.Object(players[0], "/org/mpris/MediaPlayer2"), nil
}

// parsePropertiesChanged parses a DBus PropertiesChanged signal
func parsePropertiesChanged(msg *dbus.Message) (iface string, changed map[string]dbus.Variant, ok bool) {
	if len(msg.Body) != 3 {
		return "", nil, false
	}
	iface, ok = msg.Body[0].(string)
	if !ok {
		return
	}
	changed, ok = msg.Body[1].(map[string]dbus.Variant)
	if !ok {
		return
	}
	return
}
