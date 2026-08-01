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
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/mattn/go-isatty"
	"go.elara.ws/itd/infinitime"
	"go.elara.ws/loggers"
)

var (
	firmwareUpdating = false
	// The FS must be updated when the watch is reconnected
	updateFS = false
)

var (
	cfg Config
	log *slog.Logger
)

func main() {
	showVer := flag.Bool("version", false, "Show version number and exit")
	flag.Parse()
	if *showVer {
		fmt.Println(version)
		return
	}
	
	err := loadConfig(&cfg)
	
	log = slog.New(loggers.NewPretty(os.Stderr, loggers.Options{
		Level: parseLogLevel(cfg.Logging.Level),
	}))
	
	// Defer handling the error until we have the logger set up
	if err != nil {
		log.Error("Error loading config", slog.Any("error", err))
		os.Exit(1)
	}

	// Create infinitime options struct
	opts := infinitime.Options{
		OnReconnect: func(dev *infinitime.Device) {
			if cfg.On.Reconnect.SetTime {
				// Set time to current time
				err = dev.SetTime(time.Now())
				if err != nil {
					return
				}
			}

			// If config specifies to notify on reconnect
			if cfg.On.Reconnect.Notify {
				// Send notification to InfiniTime
				err = dev.Notify("itd", "Successfully reconnected")
				if err != nil {
					return
				}
			}

			// FS must be updated on reconnect
			updateFS = true
			// Resend weather on reconnect
			sendWeatherCh <- struct{}{}
		},
	}

	ctx := context.Background()

	// Connect to InfiniTime with default options
	dev, err := infinitime.Connect(opts)
	if err != nil {
		log.Error("Error connecting to InfiniTime", slog.Any("error", err))
		os.Exit(1)
	}

	// Get firmware version
	ver, err := dev.Version()
	if err != nil {
		log.Error("Error getting firmware version", slog.Any("error", err))
		os.Exit(1)
	}

	// Log connection
	log.Info("Connected to InfiniTime", slog.String("version", ver), slog.String("addr", dev.Address()))

	// If config specifies to notify on connect
	if cfg.On.Connect.Notify {
		// Send notification to InfiniTime
		err = dev.Notify("itd", "Successfully connected")
		if err != nil {
			log.Warn("Error sending noification to InfiniTime", slog.Any("error", err))
		}
	}

	if cfg.On.Connect.SetTime {
		// Set time to current time
		err = dev.SetTime(time.Now())
		if err != nil {
			log.Warn("Error setting current time on connected InfiniTime", slog.Any("error", err))
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sig := <-sigCh
		slog.Warn("Signal received, shutting down", slog.String("signal", sig.String()))
		cancel()
	}()

	wg := WaitGroup{&sync.WaitGroup{}}

	// Initialize music controls
	err = initMusicCtrl(ctx, wg, dev)
	if err != nil {
		log.Warn("Error initializing music control", slog.Any("error", err))
	}

	// Start control socket
	err = initCallNotifs(ctx, wg, dev)
	if err != nil {
		log.Warn("Error initializing call notifications", slog.Any("error", err))
	}

	// Initialize notification relay
	err = initNotifRelay(ctx, wg, dev)
	if err != nil {
		log.Warn("Error initializing notification relay", slog.Any("error", err))
	}

	// Initializa weather
	err = initWeather(ctx, wg, dev)
	if err != nil {
		log.Warn("Error initializing weather", slog.Any("error", err))
	}

	// Initialize metrics collection
	err = initMetrics(ctx, wg, dev)
	if err != nil {
		log.Warn("Error initializing metrics collection", slog.Any("error", err))
	}

	// Initialize puremaps integration
	err = initPureMaps(ctx, wg, dev)
	if err != nil {
		log.Warn("Error initializing puremaps integration", slog.Any("error", err))
	}

	// Start fuse socket
	if cfg.Fuse.Enabled {
		err = startFUSE(ctx, wg, dev)
		if err != nil {
			log.Warn("Error starting fuse socket", slog.Any("error", err))
		}
	}

	// Start control socket
	err = startSocket(ctx, wg, dev)
	if err != nil {
		log.Warn("Error starting control socket, itctl and other clients will not work", slog.Any("error", err))
	}

	wg.Wait()
}

type x struct {
	n int
	*sync.WaitGroup
}

func (xy *x) Add(i int) {
	xy.n += i
	xy.WaitGroup.Add(i)
	fmt.Println("add: counter:", xy.n)
}

func (xy *x) Done() {
	xy.n -= 1
	xy.WaitGroup.Done()
	fmt.Println("done: counter:", xy.n)
}

func onReqPasskey() (uint32, error) {
	var out uint32
	if isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Print("Passkey: ")
		_, err := fmt.Scanln(&out)
		if err != nil {
			return 0, err
		}
	} else {
		passkey, ok, err := dlgs.Entry("Pairing", "Enter the passkey displayed on your watch.", "")
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, nil
		}
		passkeyInt, err := strconv.Atoi(passkey)
		return uint32(passkeyInt), err
	}
	return out, nil
}

func parseLogLevel(lv string) slog.Level {
	switch strings.ToLower(lv) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}