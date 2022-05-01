package api

import (
	"go.arsenm.dev/infinitime"
)

func (c *Client) FirmwareUpgrade(upgType UpgradeType, files ...string) (chan infinitime.DFUProgress, error) {
	progressCh := make(chan infinitime.DFUProgress, 5)
	err := c.client.Call(
		"ITD",
		"FirmwareUpgrade",
		FwUpgradeData{
			Type:  upgType,
			Files: files,
		},
		&progressCh,
	)
	if err != nil {
		return nil, err
	}

	return progressCh, nil
}
