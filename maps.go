package main

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/utils"
)

const (
	interfaceName     = "io.github.rinigus.PureMaps.navigator"
	iconProperty      = interfaceName + ".icon"
	narrativeProperty = interfaceName + ".narrative"
	manDistProperty   = interfaceName + ".manDist"
	progressProperty  = interfaceName + ".progress"
)

func initPureMaps(ctx context.Context, dev *infinitime.Device) error {
	// Connect to session bus. This connection is for method calls.
	conn, err := utils.NewSessionBusConn(ctx)
	if err != nil {
		return err
	}

	exists, err := pureMapsExists(ctx, conn)
	if err != nil {
		return err
	}

	// Connect to session bus. This connection is for method calls.
	monitorConn, err := utils.NewSessionBusConn(ctx)
	if err != nil {
		return err
	}

	// Define rules to listen for
	var rules = []string{
		"type='signal',interface='io.github.rinigus.PureMaps.navigator'",
	}
	var flag uint = 0
	// Becode monitor for notifications
	call := monitorConn.BusObject().CallWithContext(
		ctx, "org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, rules, flag,
	)
	if call.Err != nil {
		return call.Err
	}

	var navigator dbus.BusObject

	if exists {
		navigator = conn.Object("io.github.rinigus.PureMaps", "/io/github/rinigus/PureMaps/navigator")
		err = setAll(navigator, dev)
		if err != nil {
			log.Error().Err(err).Msg("Error setting all navigation fields")
		}
	}

	go func() {
		signalCh := make(chan *dbus.Message, 10)
		monitorConn.Eavesdrop(signalCh)

		for {
			select {
			case sig := <-signalCh:
				if sig.Type != dbus.TypeSignal {
					continue
				}

				var member string
				err = sig.Headers[dbus.FieldMember].Store(&member)
				if err != nil {
					log.Error().Err(err).Msg("Error getting dbus member field")
					continue
				}

				if !strings.HasSuffix(member, "Changed") {
					continue
				}

				log.Debug().Str("member", member).Msg("Signal received from PureMaps navigator")

				// The object must be retrieved in this loop in case PureMaps was not
				// open at the time ITD was started.
				navigator = conn.Object("io.github.rinigus.PureMaps", "/io/github/rinigus/PureMaps/navigator")
				member = strings.TrimSuffix(member, "Changed")

				switch member {
				case "icon":
					var icon string
					err = navigator.StoreProperty(iconProperty, &icon)
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error getting property")
						continue
					}

					err = dev.Navigation.SetFlag(infinitime.NavFlag(icon))
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error setting flag")
						continue
					}
				case "narrative":
					var narrative string
					err = navigator.StoreProperty(narrativeProperty, &narrative)
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error getting property")
						continue
					}

					err = dev.Navigation.SetNarrative(narrative)
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error setting flag")
						continue
					}
				case "manDist":
					var manDist string
					err = navigator.StoreProperty(manDistProperty, &manDist)
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error getting property")
						continue
					}

					err = dev.Navigation.SetManDist(manDist)
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error setting flag")
						continue
					}
				case "progress":
					var progress int32
					err = navigator.StoreProperty(progressProperty, &progress)
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error getting property")
						continue
					}

					err = dev.Navigation.SetProgress(uint8(progress))
					if err != nil {
						log.Error().Err(err).Str("property", member).Msg("Error setting flag")
						continue
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	if exists {
		log.Info().Msg("Sending PureMaps data to InfiniTime")
	}

	return nil
}

func setAll(navigator dbus.BusObject, dev *infinitime.Device) error {
	var icon string
	err := navigator.StoreProperty(iconProperty, &icon)
	if err != nil {
		return err
	}

	err = dev.Navigation.SetFlag(infinitime.NavFlag(icon))
	if err != nil {
		return err
	}

	var narrative string
	err = navigator.StoreProperty(narrativeProperty, &narrative)
	if err != nil {
		return err
	}

	err = dev.Navigation.SetNarrative(narrative)
	if err != nil {
		return err
	}

	var manDist string
	err = navigator.StoreProperty(manDistProperty, &manDist)
	if err != nil {
		return err
	}

	err = dev.Navigation.SetManDist(manDist)
	if err != nil {
		return err
	}

	var progress int32
	err = navigator.StoreProperty(progressProperty, &progress)
	if err != nil {
		return err
	}

	return dev.Navigation.SetProgress(uint8(progress))
}

// pureMapsExists checks to make sure the PureMaps service exists on the bus
func pureMapsExists(ctx context.Context, conn *dbus.Conn) (bool, error) {
	var names []string
	err := conn.BusObject().CallWithContext(
		ctx, "org.freedesktop.DBus.ListNames", 0,
	).Store(&names)
	if err != nil {
		return false, err
	}
	return strSlcContains(names, "io.github.rinigus.PureMaps"), nil
}
