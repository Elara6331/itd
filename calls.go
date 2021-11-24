package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
)

func initCallNotifs(dev *infinitime.Device) error {
	// Connect to dbus session monitorConn
	monitorConn, err := newSystemBusConn()
	if err != nil {
		return err
	}

	conn, err := newSystemBusConn()
	if err != nil {
		return err
	}

	err = monitorConn.AddMatchSignal(
		dbus.WithMatchSender("org.freedesktop.ModemManager1"),
		dbus.WithMatchInterface("org.freedesktop.ModemManager1.Modem.Voice"),
		dbus.WithMatchMember("CallAdded"),
	)
	if err != nil {
		return err
	}

	callCh := make(chan *dbus.Message, 10)
	monitorConn.Eavesdrop(callCh)
	go func() {
		for event := range callCh {
			callPath := event.Body[0].(dbus.ObjectPath)
			callObj := conn.Object("org.freedesktop.ModemManager1", callPath)

			phoneNum, err := getPhoneNum(conn, callObj)
			if err != nil {
				log.Fatal().Err(err).Send()
			}

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

func getPhoneNum(conn *dbus.Conn, callObj dbus.BusObject) (string, error) {
	var out string
	err := callObj.StoreProperty("org.freedesktop.ModemManager1.Call.Number", &out)
	if err != nil {
		return "", err
	}
	return out, nil
}

func acceptCall(conn *dbus.Conn, callObj dbus.BusObject) error {
	call := callObj.Call("org.freedesktop.ModemManager1.Call.Accept", 0)
	if call.Err != nil {
		return call.Err
	}
	return nil
}

func declineCall(conn *dbus.Conn, callObj dbus.BusObject) error {
	call := callObj.Call("org.freedesktop.ModemManager1.Call.Hangup", 0)
	if call.Err != nil {
		return call.Err
	}
	return nil
}