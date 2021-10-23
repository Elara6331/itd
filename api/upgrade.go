package api

import (
	"encoding/json"

	"go.arsenm.dev/itd/internal/types"
)

type DFUProgress types.DFUProgress

type UpgradeType uint8

const (
	UpgradeTypeArchive UpgradeType = iota
	UpgradeTypeFiles
)

func (c *Client) FirmwareUpgrade(upgType UpgradeType, files ...string) (<-chan DFUProgress, error) {
	err := json.NewEncoder(c.conn).Encode(types.Request{
		Type: types.ReqTypeFwUpgrade,
		Data: types.ReqDataFwUpgrade{
			Type:  int(upgType),
			Files: files,
		},
	})
	if err != nil {
		return nil, err
	}

	c.dfuProgressCh = make(chan DFUProgress, 5)

	return c.dfuProgressCh, nil
}
