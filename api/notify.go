package api

import "go.arsenm.dev/itd/internal/types"

func (c *Client) Notify(title string, body string) error {
	_, err := c.request(types.Request{
		Type: types.ReqTypeNotify,
		Data: types.ReqDataNotify{
			Title: title,
			Body: body,
		},
	})
	return err
}