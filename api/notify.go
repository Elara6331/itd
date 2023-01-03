package api

import (
	"context"

	"go.arsenm.dev/itd/internal/rpc"
)

func (c *Client) Notify(ctx context.Context, title, body string) error {
	_, err := c.client.Notify(ctx, &rpc.NotifyRequest{
		Title: title,
		Body:  body,
	})
	return err
}
