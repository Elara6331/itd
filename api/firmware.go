package api

import (
	"context"

	"go.elara.ws/itd/internal/rpc"
)

type DFUProgress struct {
	Sent     int64
	Received int64
	Total    int64
	Err      error
}

func (c *Client) FirmwareUpgrade(ctx context.Context, upgType UpgradeType, files ...string) (chan DFUProgress, error) {
	progressCh := make(chan DFUProgress, 5)
	fc, err := c.client.FirmwareUpgrade(ctx, &rpc.FirmwareUpgradeRequest{
		Type:  rpc.FirmwareUpgradeRequest_Type(upgType),
		Files: files,
	})
	if err != nil {
		return nil, err
	}

	go fsRecvToChannel[rpc.DFUProgress](fc, progressCh, func(evt *rpc.DFUProgress, err error) DFUProgress {
		return DFUProgress{
			Sent:     evt.Sent,
			Received: evt.Recieved,
			Total:    evt.Total,
			Err:      err,
		}
	})

	return progressCh, nil
}
