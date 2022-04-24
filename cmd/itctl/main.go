package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"go.arsenm.dev/itd/api"
)

var client *api.Client

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := cli.App{
		Name: "itctl",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "socket-path",
				Aliases: []string{"s"},
				Value:   api.DefaultAddr,
				Usage:   "Path to itd socket",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "filesystem",
				Aliases: []string{"fs"},
				Usage:   "Perform filesystem operations on the PineTime",
				Subcommands: []*cli.Command{
					{
						Name:      "list",
						ArgsUsage: "[dir]",
						Aliases:   []string{"ls"},
						Usage:     "List a directory",
						Action:    fsList,
					},
					{
						Name:      "mkdir",
						ArgsUsage: "<paths...>",
						Usage:     "Create new directories",
						Action:    fsMkdir,
					},
					{
						Name:      "move",
						ArgsUsage: "<old> <new>",
						Aliases:   []string{"mv"},
						Usage:     "Move a file or directory",
						Action:    fsMove,
					},
					{
						Name:        "read",
						ArgsUsage:   `<remote path> <local path>`,
						Usage:       "Read a file from InfiniTime.",
						Description: `Read is used to read files from InfiniTime's filesystem. A "-" can be used to signify stdout`,
						Action:      fsRead,
					},
					{
						Name:      "remove",
						ArgsUsage: "<paths...>",
						Aliases:   []string{"rm"},
						Usage:     "Remove a file from InfiniTime",
						Action:    fsRemove,
					},
					{
						Name:        "write",
						ArgsUsage:   `<local path> <remote path>`,
						Usage:       "Write a file to InfiniTime",
						Description: `Write is used to write files to InfiniTime's filesystem. A "-" can be used to signify stdin`,
						Action:      fsWrite,
					},
				},
			},
			{
				Name:    "firmware",
				Aliases: []string{"fw"},
				Usage:   "Manage InfiniTime firmware",
				Subcommands: []*cli.Command{
					{
						Flags: []cli.Flag{
							&cli.PathFlag{
								Name:    "init-packet",
								Aliases: []string{"i"},
								Usage:   "Path to init packet (.dat file)",
							},
							&cli.PathFlag{
								Name:    "firmware",
								Aliases: []string{"f"},
								Usage:   "Path to firmware image (.bin file)",
							},
							&cli.PathFlag{
								Name:    "archive",
								Aliases: []string{"a"},
								Usage:   "Path to firmware archive (.zip file)",
							},
						},
						Name:    "upgrade",
						Aliases: []string{"upg"},
						Usage:   "Upgrade InfiniTime firmware using files or archive",
						Action:  fwUpgrade,
					},
					{
						Name:    "version",
						Aliases: []string{"ver"},
						Usage:   "Get firmware version of InfiniTime",
						Action:  fwVersion,
					},
				},
			},
			{
				Name:  "get",
				Usage: "Get information from InfiniTime",
				Subcommands: []*cli.Command{
					{
						Name:    "address",
						Aliases: []string{"addr"},
						Usage:   "Get InfiniTime's bluetooth address",
						Action:  getAddress,
					},
					{
						Name:    "battery",
						Aliases: []string{"batt"},
						Usage:   "Get InfiniTime's battery percentage",
						Action:  getBattery,
					},
					{
						Name:   "heart",
						Usage:  "Get heart rate from InfiniTime",
						Action: getHeart,
					},
					{
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "shell"},
						},
						Name:   "motion",
						Usage:  "Get motion values from InfiniTime",
						Action: getMotion,
					},
					{
						Name:   "steps",
						Usage:  "Get step count from InfiniTime",
						Action: getSteps,
					},
				},
			},
			{
				Name:   "notify",
				Usage:  "Send notification to InfiniTime",
				Action: notify,
			},
			{
				Name:  "set",
				Usage: "Set information on InfiniTime",
				Subcommands: []*cli.Command{
					{
						Name:      "time",
						ArgsUsage: `<ISO8601|"now">`,
						Usage:     "Set InfiniTime's clock to specified time",
						Action:    setTime,
					},
				},
			},
			{
				Name:    "update",
				Usage:   "Update information on InfiniTime",
				Aliases: []string{"upd"},
				Subcommands: []*cli.Command{
					{
						Name:   "weather",
						Usage:  "Force an immediate update of weather data",
						Action: updateWeather,
					},
				},
			},
			{
				Name:  "watch",
				Usage: "Watch a value for changes",
				Subcommands: []*cli.Command{
					{
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "json"},
							&cli.BoolFlag{Name: "shell"},
						},
						Name:   "heart",
						Usage:  "Watch heart rate value for changes",
						Action: watchHeart,
					},
					{
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "json"},
							&cli.BoolFlag{Name: "shell"},
						},
						Name:   "steps",
						Usage:  "Watch step count value for changes",
						Action: watchStepCount,
					},
					{
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "json"},
							&cli.BoolFlag{Name: "shell"},
						},
						Name:   "motion",
						Usage:  "Watch motion coordinates for changes",
						Action: watchMotion,
					},
					{
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "json"},
							&cli.BoolFlag{Name: "shell"},
						},
						Name:    "battery",
						Aliases: []string{"batt"},
						Usage:   "Watch battery level value for changes",
						Action:  watchBattLevel,
					},
				},
			},
		},
		Before: func(c *cli.Context) error {
			newClient, err := api.New(c.String("socket-path"))
			if err != nil {
				return err
			}
			client = newClient
			return nil
		},
		After: func(*cli.Context) error {
			if client != nil {
				client.Close()
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while running app")
	}
}

func catchSignal(fn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(
		sigCh,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	go func() {
		<-sigCh
		fn()
		os.Exit(0)
	}()
}
