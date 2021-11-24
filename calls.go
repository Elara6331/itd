package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
)

func initCallNotifs(dev *infinitime.Device) error {
	// Connect to system bus. This connection is for monitoring.
	monitorConn, err := newSystemBusConn()
	if err != nil {
		return err
	}

	// Connect to system bus. This connection is for method calls.
	conn, err := newSystemBusConn()
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

	go func() {
		// For every message received
		for event := range callCh {
			// Get path to call object
			callPath := event.Body[0].(dbus.ObjectPath)
			// Get call object
			callObj := conn.Object("org.freedesktop.ModemManager1", callPath)

			// Get phone number from call object using method call connection
			phoneNum, err := getPhoneNum(conn, callObj)
			if err != nil {
				log.Fatal().Err(err).Send()
			}

			// Send call notification to InfiniTime
			resCh, err := dev.NotifyCall(phoneNum)
			if err != nil {
				continue
			}

			go func() {
				// Wait for PineTime response
				res := <-resCh
				switch res {
				case infinitime.CallStatusAccepted:
					// Attempt to accept call
					err = acceptCall(conn, callObj)
					if err != nil {
						log.Warn().Err(err).Msg("Error accepting call")
					}
				case infinitime.CallStatusDeclined:
					// Attempt to decline call
					err = declineCall(conn, callObj)
					if err != nil {
						log.Warn().Err(err).Msg("Error declining call")
					}
				case infinitime.CallStatusMuted:
					// Warn about unimplemented muting
					log.Warn().Msg("Muting calls is not implemented")
				}
			}()
		}
	}()
	return nil
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
func acceptCall(conn *dbus.Conn, callObj dbus.BusObject) error {
	// Call Accept() method on DBus object
	call := callObj.Call("org.freedesktop.ModemManager1.Call.Accept", 0)
	if call.Err != nil {
		return call.Err
	}
	return nil
}

// getPhoneNum declines a call using a DBus connection
func declineCall(conn *dbus.Conn, callObj dbus.BusObject) error {
	// Call Hangup() method on DBus object
	call := callObj.Call("org.freedesktop.ModemManager1.Call.Hangup", 0)
	if call.Err != nil {
		return call.Err
	}
	return nil
}
