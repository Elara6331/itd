package utils

import (
	"context"

	"github.com/godbus/dbus/v5"
)

func NewSystemBusConn(ctx context.Context) (*dbus.Conn, error) {
	// Connect to dbus session bus
	conn, err := dbus.SystemBusPrivate(dbus.WithContext(ctx))
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

func NewSessionBusConn(ctx context.Context) (*dbus.Conn, error) {
	// Connect to dbus session bus
	conn, err := dbus.SessionBusPrivate(dbus.WithContext(ctx))
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
