package api

import (
	"context"

	"go.elara.ws/itd/internal/rpc"
)

func (c *Client) WeatherUpdate(ctx context.Context) error {
	_, err := c.client.WeatherUpdate(ctx, &rpc.Empty{})
	return err
}
