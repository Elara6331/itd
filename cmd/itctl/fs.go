package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/urfave/cli/v2"
)

func fsList(c *cli.Context) error {
	dirPath := "/"
	if c.Args().Len() > 0 {
		dirPath = c.Args().Get(0)
	}

	listing, err := client.FS().ReadDir(c.Context, dirPath)
	if err != nil {
		return err
	}

	for _, entry := range listing {
		fmt.Println(entry)
	}

	return nil
}

func fsMkdir(c *cli.Context) error {
	if c.Args().Len() < 1 {
		return cli.Exit("Command mkdir requires one or more arguments", 1)
	}

	var err error
	if c.Bool("parents") {
		err = client.FS().MkdirAll(c.Context, c.Args().Slice()...)
	} else {
		err = client.FS().Mkdir(c.Context, c.Args().Slice()...)
	}
	if err != nil {
		return err
	}

	return nil
}

func fsMove(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return cli.Exit("Command move requires two arguments", 1)
	}

	err := client.FS().Rename(c.Context, c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return err
	}

	return nil
}

func fsRead(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return cli.Exit("Command read requires two arguments", 1)
	}

	var tmpFile *os.File
	var path string
	var err error
	if c.Args().Get(1) == "-" {
		tmpFile, err = os.CreateTemp("/tmp", "itctl.*")
		if err != nil {
			return err
		}
		path = tmpFile.Name()
	} else {
		path, err = filepath.Abs(c.Args().Get(1))
		if err != nil {
			return err
		}
	}

	progress, err := client.FS().Download(c.Context, path, c.Args().Get(0))
	if err != nil {
		return err
	}

	// Create progress bar template
	barTmpl := `{{counters . }} B {{bar . "|" "-" (cycle .) " " "|"}} {{percent . }} {{rtime . "%s"}}`
	// Start full bar at 0 total
	bar := pb.ProgressBarTemplate(barTmpl).Start(0)
	// Get progress events
	for event := range progress {
		if event.Err != nil {
			return event.Err
		}

		// Set total bytes in progress bar
		bar.SetTotal(int64(event.Total))
		// Set amount of bytes sent in progress bar
		bar.SetCurrent(int64(event.Sent))
	}
	bar.Finish()

	if c.Args().Get(1) == "-" {
		io.Copy(os.Stdout, tmpFile)
		os.Stdout.WriteString("\n")
		os.Stdout.Sync()
		tmpFile.Close()
	}

	return nil
}

func fsRemove(c *cli.Context) error {
	if c.Args().Len() < 1 {
		return cli.Exit("Command remove requires one or more arguments", 1)
	}

	var err error
	if c.Bool("recursive") {
		err = client.FS().RemoveAll(c.Context, c.Args().Slice()...)
	} else {
		err = client.FS().Remove(c.Context, c.Args().Slice()...)
	}
	if err != nil {
		return err
	}

	return nil
}

func fsWrite(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return cli.Exit("Command write requires two arguments", 1)
	}

	var tmpFile *os.File
	var path string
	var err error
	if c.Args().Get(0) == "-" {
		tmpFile, err = os.CreateTemp("/tmp", "itctl.*")
		if err != nil {
			return err
		}
		path = tmpFile.Name()
	} else {
		path, err = filepath.Abs(c.Args().Get(0))
		if err != nil {
			return err
		}
	}

	if c.Args().Get(0) == "-" {
		io.Copy(tmpFile, os.Stdin)
		defer tmpFile.Close()
		defer os.Remove(path)
	}

	progress, err := client.FS().Upload(c.Context, c.Args().Get(1), path)
	if err != nil {
		return err
	}

	// Create progress bar template
	barTmpl := `{{counters . }} B {{bar . "|" "-" (cycle .) " " "|"}} {{percent . }} {{rtime . "%s"}}`
	// Start full bar at 0 total
	bar := pb.ProgressBarTemplate(barTmpl).Start(0)
	// Get progress events
	for event := range progress {
		if event.Err != nil {
			return event.Err
		}

		// Set total bytes in progress bar
		bar.SetTotal(int64(event.Total))
		// Set amount of bytes sent in progress bar
		bar.SetCurrent(int64(event.Sent))
	}

	return nil
}
