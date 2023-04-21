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
	"os"
	"os/signal"
	"strconv"
	"sync"

	"syscall"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/knadh/koanf"
	"github.com/mattn/go-isatty"
	"go.elara.ws/infinitime"
	"go.elara.ws/logger"
	"go.elara.ws/logger/log"
)

var k = koanf.New(".")

var (
	firmwareUpdating = false
	// The FS must be updated when the watch is reconnected
	updateFS = false
)

func main() {
	showVer := flag.Bool("version", false, "Show version number and exit")
	flag.Parse()
	// If version requested, print and exit
	if *showVer {
		fmt.Println(version)
		return
	}

	level, err := logger.ParseLogLevel(k.String("logging.level"))
	if err != nil {
		level = logger.LogLevelInfo
	}

	// Initialize infinitime library
	infinitime.Init(k.String("bluetooth.adapter"))
	// Cleanly exit after function
	defer infinitime.Exit()

	// Create infinitime options struct
	opts := &infinitime.Options{
		AttemptReconnect: k.Bool("conn.reconnect"),
		WhitelistEnabled: k.Bool("conn.whitelist.enabled"),
		Whitelist:        k.Strings("conn.whitelist.devices"),
		OnReqPasskey:     onReqPasskey,
		Logger:           log.Logger,
		LogLevel:         level,
	}

	ctx := context.Background()

	// Connect to InfiniTime with default options
	dev, err := infinitime.Connect(ctx, opts)
	if err != nil {
		log.Fatal("Error connecting to InfiniTime").Err(err).Send()
	}

	// When InfiniTime reconnects
	opts.OnReconnect = func() {
		if k.Bool("on.reconnect.setTime") {
			// Set time to current time
			err = dev.SetTime(time.Now())
			if err != nil {
				return
			}
		}

		// If config specifies to notify on reconnect
		if k.Bool("on.reconnect.notify") {
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
	}

	// Get firmware version
	ver, err := dev.Version()
	if err != nil {
		log.Error("Error getting firmware version").Err(err).Send()
	}

	// Log connection
	log.Info("Connected to InfiniTime").Str("version", ver).Send()

	// If config specifies to notify on connect
	if k.Bool("on.connect.notify") {
		// Send notification to InfiniTime
		err = dev.Notify("itd", "Successfully connected")
		if err != nil {
			log.Error("Error sending notification to InfiniTime").Err(err).Send()
		}
	}

	// Set time to current time
	err = dev.SetTime(time.Now())
	if err != nil {
		log.Error("Error setting current time on connected InfiniTime").Err(err).Send()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sig := <-sigCh
		log.Warn("Signal received, shutting down").Stringer("signal", sig).Send()
		cancel()
	}()

	wg := WaitGroup{&sync.WaitGroup{}}

	// Initialize music controls
	err = initMusicCtrl(ctx, wg, dev)
	if err != nil {
		log.Error("Error initializing music control").Err(err).Send()
	}

	// Start control socket
	err = initCallNotifs(ctx, wg, dev)
	if err != nil {
		log.Error("Error initializing call notifications").Err(err).Send()
	}

	// Initialize notification relay
	err = initNotifRelay(ctx, wg, dev)
	if err != nil {
		log.Error("Error initializing notification relay").Err(err).Send()
	}

	// Initializa weather
	err = initWeather(ctx, wg, dev)
	if err != nil {
		log.Error("Error initializing weather").Err(err).Send()
	}

	// Initialize metrics collection
	err = initMetrics(ctx, wg, dev)
	if err != nil {
		log.Error("Error intializing metrics collection").Err(err).Send()
	}

	// Initialize puremaps integration
	err = initPureMaps(ctx, wg, dev)
	if err != nil {
		log.Error("Error intializing puremaps integration").Err(err).Send()
	}

	// Start fuse socket
	if k.Bool("fuse.enabled") {
		err = startFUSE(ctx, wg, dev)
		if err != nil {
			log.Error("Error starting fuse socket").Err(err).Send()
		}
	}

	// Start control socket
	err = startSocket(ctx, wg, dev)
	if err != nil {
		log.Error("Error starting socket").Err(err).Send()
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
