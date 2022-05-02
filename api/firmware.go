package api

import (
	"context"

	"go.arsenm.dev/infinitime"
)

func (c *Client) FirmwareUpgrade(ctx context.Context, upgType UpgradeType, files ...string) (chan infinitime.DFUProgress, error) {
	progressCh := make(chan infinitime.DFUProgress, 5)
	err := c.client.Call(
		ctx,
		"ITD",
		"FirmwareUpgrade",
		FwUpgradeData{
			Type:  upgType,
			Files: files,
		},
		progressCh,
	)
	if err != nil {
		return nil, err
	}

	return progressCh, nil
}
