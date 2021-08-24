package main

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	// Set up logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set config settings
	setCfgDefaults()
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("/etc")
	viper.SetConfigName("itd")
	viper.SetConfigType("toml")
	viper.WatchConfig()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("itd")
	// Ignore error because defaults set
	viper.ReadInConfig()
	viper.AutomaticEnv()
}

func setCfgDefaults() {
	viper.SetDefault("cfg.version", 2)

	viper.SetDefault("socket.path", "/tmp/itd/socket")

	viper.SetDefault("conn.reconnect", true)

	viper.SetDefault("on.connect.notify", true)

	viper.SetDefault("on.reconnect.notify", true)
	viper.SetDefault("on.reconnect.setTime", true)

	viper.SetDefault("notifs.ignore.sender", []string{})
	viper.SetDefault("notifs.ignore.summary", []string{"InfiniTime"})
	viper.SetDefault("notifs.ignore.body", []string{})

	viper.SetDefault("music.vol.interval", 5)
}
