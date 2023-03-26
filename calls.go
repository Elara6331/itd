package main

import (
	"context"
	"sync"

	"github.com/godbus/dbus/v5"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/utils"
	"go.arsenm.dev/logger/log"
)

func initCallNotifs(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	// Connect to system bus. This connection is for method calls.
	conn, err := utils.NewSystemBusConn(ctx)
	if err != nil {
		return err
	}

	// Check if modem manager interface exists
	exists, err := modemManagerExists(ctx, conn)
	if err != nil {
		return err
	}

	// If it does not exist, stop function
	if !exists {
		conn.Close()
		return nil
	}

	// Connect to system bus. This connection is for monitoring.
	monitorConn, err := utils.NewSystemBusConn(ctx)
	if err != nil {
		return err
	}

	// Add match for new calls to monitor connection
	err = monitorConn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.ModemManager1"),
		dbus.WithMatchInterface("org.freedesktop.ModemManager1.Modem.Voice"),
		dbus.WithMatchMember("CallAdded"),
	)
	if err != nil {
		return err
	}

	// Create channel to receive calls
	callCh := make(chan *dbus.Message, 5)
	// Notify channel upon received message
	monitorConn.Eavesdrop(callCh)

	var respHandlerOnce sync.Once
	var callObj dbus.BusObject

	wg.Add(1)
	go func() {
		defer wg.Done("callNotifs")
		for {
			select {
			case event := <-callCh:
				// Get path to call object
				callPath := event.Body[0].(dbus.ObjectPath)
				// Get call object
				callObj = conn.Object("org.freedesktop.ModemManager1", callPath)

				// Get phone number from call object using method call connection
				phoneNum, err := getPhoneNum(conn, callObj)
				if err != nil {
					log.Error("Error getting phone number").Err(err).Send()
					continue
				}

				// Send call notification to InfiniTime
				resCh, err := dev.NotifyCall(phoneNum)
				if err != nil {
					continue
				}

				go respHandlerOnce.Do(func() {
					// Wait for PineTime response
					for res := range resCh {
						switch res {
						case infinitime.CallStatusAccepted:
							// Attempt to accept call
							err = acceptCall(ctx, conn, callObj)
							if err != nil {
								log.Warn("Error accepting call").Err(err).Send()
							}
						case infinitime.CallStatusDeclined:
							// Attempt to decline call
							err = declineCall(ctx, conn, callObj)
							if err != nil {
								log.Warn("Error declining call").Err(err).Send()
							}
						case infinitime.CallStatusMuted:
							// Warn about unimplemented muting
							log.Warn("Muting calls is not implemented").Send()
						}
					}
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Info("Relaying calls to InfiniTime").Send()
	return nil
}

func modemManagerExists(ctx context.Context, conn *dbus.Conn) (bool, error) {
	var names []string
	err := conn.BusObject().CallWithContext(
		ctx, "org.freedesktop.DBus.ListNames", 0,
	).Store(&names)
	if err != nil {
		return false, err
	}
	return strSlcContains(names, "org.freedesktop.ModemManager1"), nil
}

// getPhoneNum gets a phone number from a call object using a DBus connection
func getPhoneNum(conn *dbus.Conn, callObj dbus.BusObject) (string, error) {
	var out string
	// Get number property on DBus object and store return value in out
	err := callObj.StoreProperty("org.freedesktop.ModemManager1.Call.Number", &out)
	if err != nil {
		return "", err
	}
	return out, nil
}

// getPhoneNum accepts a call using a DBus connection
func acceptCall(ctx context.Context, conn *dbus.Conn, callObj dbus.BusObject) error {
	// Call Accept() method on DBus object
	call := callObj.CallWithContext(
		ctx, "org.freedesktop.ModemManager1.Call.Accept", 0,
	)
	if call.Err != nil {
		return call.Err
	}
	return nil
}

// getPhoneNum declines a call using a DBus connection
func declineCall(ctx context.Context, conn *dbus.Conn, callObj dbus.BusObject) error {
	// Call Hangup() method on DBus object
	call := callObj.CallWithContext(
		ctx, "org.freedesktop.ModemManager1.Call.Hangup", 0,
	)
	if call.Err != nil {
		return call.Err
	}
	return nil
}
