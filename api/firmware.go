package api

import (
	"context"
	"time"

	"go.arsenm.dev/infinitime"
)

func (c *Client) FirmwareUpgrade(upgType UpgradeType, files ...string) (chan infinitime.DFUProgress, error) {
	var id string
	err := c.itdClient.Call(
		context.Background(),
		"FirmwareUpgrade",
		FwUpgradeData{
			Type:  upgType,
			Files: files,
		},
		&id,
	)
	if err != nil {
		return nil, err
	}

	progressCh := make(chan infinitime.DFUProgress, 5)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			progressCh <- val.(infinitime.DFUProgress)
		}
		close(progressCh)
	}()

	return progressCh, nil
}
