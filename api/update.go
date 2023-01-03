package api

import (
	"context"

	"go.arsenm.dev/itd/internal/rpc"
)

func (c *Client) WeatherUpdate(ctx context.Context) error {
	_, err := c.client.WeatherUpdate(ctx, &rpc.Empty{})
	return err
}
