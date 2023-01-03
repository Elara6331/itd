package main

import (
	"context"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/urfave/cli/v2"
	"go.arsenm.dev/infinitime"
)

func resourcesLoad(c *cli.Context) error {
	return resLoad(c.Context, c.Args().Slice())
}

func resLoad(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return cli.Exit("Command load requires one argument.", 1)
	}

	// Create progress bar templates
	rmTmpl := `Removing {{string . "filename"}}`
	upTmpl := `Uploading {{string . "filename"}} {{counters . }} B {{bar . "|" "-" (cycle .) " " "|"}} {{percent . }} {{rtime . "%s"}}`
	// Start full bar at 0 total
	bar := pb.ProgressBarTemplate(rmTmpl).Start(0)

	path, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	progCh, err := client.FS().LoadResources(ctx, path)
	if err != nil {
		return err
	}

	for evt := range progCh {
		if evt.Err != nil {
			return evt.Err
		}

		if evt.Operation == infinitime.ResourceOperationRemoveObsolete {
			bar.SetTemplateString(rmTmpl)
			bar.Set("filename", evt.Name)
		} else {
			bar.SetTemplateString(upTmpl)
			bar.Set("filename", evt.Name)

			bar.SetTotal(evt.Total)
			bar.SetCurrent(evt.Sent)
		}
	}

	bar.Finish()

	return nil
}
