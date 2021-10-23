package api

import (
	"encoding/json"

	"go.arsenm.dev/itd/internal/types"
)

// DFUProgress stores the progress of a DFU upfate
type DFUProgress types.DFUProgress

// UpgradeType indicates the type of upgrade to be performed
type UpgradeType uint8

// Type of DFU upgrade
const (
	UpgradeTypeArchive UpgradeType = iota
	UpgradeTypeFiles
)

// FirmwareUpgrade initiates a DFU update and returns the progress channel
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
