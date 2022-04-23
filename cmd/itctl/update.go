package main

import "github.com/urfave/cli/v2"

func updateWeather(c *cli.Context) error {
	return client.WeatherUpdate()
}
