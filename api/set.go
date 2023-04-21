package api

import (
	"context"
	"time"

	"go.elara.ws/itd/internal/rpc"
)

func (c *Client) SetTime(ctx context.Context, t time.Time) error {
	_, err := c.client.SetTime(ctx, &rpc.SetTimeRequest{UnixNano: t.UnixNano()})
	return err
}
