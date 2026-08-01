package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func getCfgPath() (string, error) {
	userCfgDir, err := os.UserConfigDir()
	return filepath.Join(userCfgDir, "itd", "itd.toml"), err
}

func Load(cfg *Config) error {
	*cfg = defaults

	cfgPath, err := getCfgPath()
	if err != nil {
		return err
	}

	cfgDir := filepath.Dir(cfgPath)
	if _, err = os.ReadDir(cfgDir); err != nil {
		err = os.MkdirAll(cfgDir, 0o700)
		if err != nil {
			return err
		}
	}
	(*cfg).Dir = cfgDir

	fl, err := os.Open(cfgPath)
	if err != nil {
		return nil // cfg is already set to defaults
	}

	return toml.NewDecoder(fl).Decode(cfg)
}

func getRuntimeDir() string {
	if xrd := os.Getenv("XDG_RUNTIME_DIR"); xrd != "" {
		return xrd
	}
	return fmt.Sprintf("/run/user/%d", os.Getuid())
}

var defaults = Config{
	Bluetooh: Bluetooh{Adapter: "hci0"},
	Socket:   Socket{Path: filepath.Join(getRuntimeDir(), "itd.sock")},
	Conn: Conn{
		Reconnect: true,
		Whitelist: Whitelist{Enabled: false},
	},
	On: On{
		Connect:   Hook{Notify: true, SetTime: true},
		Reconnect: Hook{Notify: true, SetTime: true},
	},
	Notifs: Notifs{
		Translit: NotifsTranslit{
			Use: []string{"eASCII"},
		},
		Ignore: NotifsIgnore{
			Summary: []string{"InfiniTime"},
		},
	},
	Music: Music{
		Vol: Volume{Interval: 5},
	},
	Fuse: Fuse{
		Enabled:    false,
		Mountpoint: "/tmp/itd/mnt",
	},
}

type Config struct {
	Dir      string   `toml:"-"`
	Logging  Logging  `toml:"logging"`
	Weather  Weather  `toml:"weather"`
	Bluetooh Bluetooh `toml:"bluetooth"`
	Notifs   Notifs   `toml:"notifs"`
	Conn     Conn     `toml:"conn"`
	On       On       `toml:"on"`
	Fuse     Fuse     `toml:"fuse"`
	Music    Music    `toml:"music"`
	Metrics  Metrics  `toml:"metrics"`
	Socket   Socket   `toml:"socket"`
}

type Weather struct {
	Enabled  bool   `toml:"enabled"`
	Location string `toml:"location"`
}

type Logging struct {
	Level string `toml:"level"`
}

type Bluetooh struct {
	Adapter string `toml:"adapter"`
}

type Socket struct {
	Path string `toml:"path"`
}

type Metrics struct {
	Enabled   bool   `toml:"enabled"`
	HeartRate Metric `toml:"heartRate"`
	StepCount Metric `toml:"stepCount"`
	BattLevel Metric `toml:"battLevel"`
	Motion    Metric `toml:"motion"`
}

type Metric struct {
	Enabled bool `toml:"enabled"`
}

type Conn struct {
	Reconnect bool      `toml:"reconnect"`
	Whitelist Whitelist `toml:"whitelist"`
}

type On struct {
	Connect   Hook `toml:"connect"`
	Reconnect Hook `toml:"reconnect"`
}

type Hook struct {
	Notify  bool `toml:"notify"`
	SetTime bool `toml:"setTime"`
}

type Whitelist struct {
	Enabled bool     `toml:"enabled"`
	Devices []string `toml:"devices"`
}

type Notifs struct {
	Translit NotifsTranslit `toml:"translit"`
	Ignore   NotifsIgnore   `toml:"ignore"`
}

type NotifsTranslit struct {
	Use    []string `toml:"user"`
	Custom []string `toml:"custom"`
}

type NotifsIgnore struct {
	Sender  []string `toml:"sender"`
	Summary []string `toml:"summary"`
	Body    []string `toml:"body"`
}

type Fuse struct {
	Enabled    bool   `toml:"enabled"`
	Mountpoint string `toml:"mountpoint"`
}

type Music struct {
	Vol Volume `toml:"vol"`
}

type Volume struct {
	Interval uint `toml:"interval"`
}
