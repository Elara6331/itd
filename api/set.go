package api

import (
	"context"
	"time"
)

func (c *Client) SetTime(ctx context.Context, t time.Time) error {
	return c.client.Call(
		ctx,
		"ITD",
		"SetTime",
		t,
		nil,
	)
}
