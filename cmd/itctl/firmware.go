package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/urfave/cli/v2"
	"go.arsenm.dev/itd/api"
)

func fwUpgrade(c *cli.Context) error {
	start := time.Now()

	var upgType api.UpgradeType
	var files []string
	// Get relevant data struct
	if c.String("archive") != "" {
		// Get archive data struct
		upgType = api.UpgradeTypeArchive
		files = []string{c.String("archive")}
	} else if c.String("init-packet") != "" && c.String("firmware") != "" {
		// Get files data struct
		upgType = api.UpgradeTypeFiles
		files = []string{c.String("init-packet"), c.String("firmware")}
	} else {
		return cli.Exit("Upgrade command requires either archive or init packet and firmware.", 1)
	}

	progress, err := client.FirmwareUpgrade(upgType, abs(files)...)
	if err != nil {
		return err
	}

	// Create progress bar template
	barTmpl := `{{counters . }} B {{bar . "|" "-" (cycle .) " " "|"}} {{percent . }} {{rtime . "%s"}}`
	// Start full bar at 0 total
	bar := pb.ProgressBarTemplate(barTmpl).Start(0)
	// Create new scanner of connection
	for event := range progress {
		// Set total bytes in progress bar
		bar.SetTotal(event.Total)
		// Set amount of bytes received in progress bar
		bar.SetCurrent(int64(event.Received))
		// If transfer finished, break
		if int64(event.Sent) == event.Total {
			break
		}
	}
	// Finish progress bar
	bar.Finish()

	fmt.Printf("Transferred %d B in %s.\n", bar.Total(), time.Since(start))
	fmt.Println("Remember to validate the new firmware in the InfiniTime settings.")

	return nil
}

func fwVersion(c *cli.Context) error {
	version, err := client.Version()
	if err != nil {
		return err
	}

	fmt.Println(version)
	return nil
}

func abs(paths []string) []string {
	for index, path := range paths {
		newPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		paths[index] = newPath
	}
	return paths
}
