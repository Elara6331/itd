package api

import "context"

func (c *Client) WeatherUpdate(ctx context.Context) error {
	return c.client.Call(
		ctx,
		"ITD",
		"WeatherUpdate",
		nil,
		nil,
	)
}
