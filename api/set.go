package api

import (
	"time"
)

func (c *Client) SetTime(t time.Time) error {
	return c.client.Call(
		"ITD",
		"SetTime",
		t,
		nil,
	)
}
