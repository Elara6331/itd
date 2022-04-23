package api

import "context"

func (c *Client) WeatherUpdate() error {
	return c.itdClient.Call(
		context.Background(),
		"WeatherUpdate",
		nil,
		nil,
	)
}