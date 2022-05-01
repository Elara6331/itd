package api

func (c *Client) WeatherUpdate() error {
	return c.client.Call(
		"ITD",
		"WeatherUpdate",
		nil,
		nil,
	)
}
