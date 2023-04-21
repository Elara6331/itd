package api

import (
	"context"

	"go.elara.ws/itd/internal/rpc"
)

func (c *Client) Notify(ctx context.Context, title, body string) error {
	_, err := c.client.Notify(ctx, &rpc.NotifyRequest{
		Title: title,
		Body:  body,
	})
	return err
}
