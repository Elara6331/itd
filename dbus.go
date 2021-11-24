package main

import "github.com/godbus/dbus/v5"

func newSystemBusConn() (*dbus.Conn, error) {
	// Connect to dbus session bus
	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}
	err = conn.Auth(nil)
	if err != nil {
		return nil, err
	}
	err = conn.Hello()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newSessionBusConn() (*dbus.Conn, error) {
	// Connect to dbus session bus
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, err
	}
	err = conn.Auth(nil)
	if err != nil {
		return nil, err
	}
	err = conn.Hello()
	if err != nil {
		return nil, err
	}
	return conn, nil
}
