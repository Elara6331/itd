package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Set up logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get user's configuration directory
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	// Set config defaults
	setCfgDefaults()

	// Load config files
	etcProvider := file.Provider("/etc/itd.toml")
	cfgProvider := file.Provider(filepath.Join(cfgDir, "itd.toml"))
	k.Load(etcProvider, toml.Parser())
	k.Load(cfgProvider, toml.Parser())

	// Watch configs for changes
	cfgWatch(etcProvider)
	cfgWatch(cfgProvider)

	// Load envireonment variables
	k.Load(env.Provider("ITD_", "_", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "ITD_"))
	}), nil)
}

func cfgWatch(provider *file.File) {
	// Watch for changes and reload when detected
	provider.Watch(func(_ interface{}, err error) {
		if err != nil {
			return
		}

		k.Load(provider, toml.Parser())
	})
}

func setCfgDefaults() {
	k.Load(confmap.Provider(map[string]interface{}{
		"socket.path": "/tmp/itd/socket",

		"conn.reconnect": true,

		"conn.whitelist.enabled": false,
		"conn.whitelist.devices": []string{},

		"on.connect.notify": true,

		"on.reconnect.notify":  true,
		"on.reconnect.setTime": true,

		"notifs.translit.use":    []string{"eASCII"},
		"notifs.translit.custom": []string{},

		"notifs.ignore.sender":  []string{},
		"notifs.ignore.summary": []string{"InfiniTime"},
		"notifs.ignore.body":    []string{},

		"music.vol.interval": 5,
	}, "."), nil)
}
