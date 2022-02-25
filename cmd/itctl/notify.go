package main

import "github.com/urfave/cli/v2"

func notify(c *cli.Context) error {
	// Ensure required arguments
	if c.Args().Len() != 2 {
		return cli.Exit("Command notify requires two arguments", 1)
	}

	err := client.Notify(c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}

	return nil
}
