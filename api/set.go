package api

import (
	"context"
	"time"
)

func (c *Client) SetTime(t time.Time) error {
	return c.itdClient.Call(
		context.Background(),
		"SetTime",
		t,
		nil,
	)
}
