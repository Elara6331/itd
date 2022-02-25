package main

import (
	"time"

	"github.com/urfave/cli/v2"
)

func setTime(c *cli.Context) error {
	// Ensure required arguments
	if c.Args().Len() < 1 {
		return cli.Exit("Command time requires one argument", 1)
	}

	if c.Args().Get(0) == "now" {
		return client.SetTimeNow()
	} else {
		parsed, err := time.Parse(time.RFC3339, c.Args().Get(0))
		if err != nil {
			return err
		}
		return client.SetTime(parsed)
	}
}
