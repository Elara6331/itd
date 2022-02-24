package api

import (
	"go.arsenm.dev/itd/internal/types"
)

// UpdateWeather sends the update weather signal,
// immediately sending current weather data
func (c *Client) UpdateWeather() error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeWeatherUpdate,
	})
	if err != nil {
		return err
	}
	return nil
}
