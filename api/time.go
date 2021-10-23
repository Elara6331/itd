package api

import (
	"time"

	"go.arsenm.dev/itd/internal/types"
)

// SetTime sets the given time on the connected device
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

// SetTimeNow sets the time on the connected device to
// the current time. This is more accurate than
// SetTime(time.Now()) due to RFC3339 formatting
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
