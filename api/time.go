package api

import (
	"time"

	"go.arsenm.dev/itd/internal/types"
)

func (c *Client) SetTime(t time.Time) error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeSetTime,
		Data: t.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SetTimeNow() error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeSetTime,
		Data: "now",
	})
	if err != nil {
		return err
	}
	return nil
}
