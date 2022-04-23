package api

import (
	"context"
)

func (c *Client) Notify(title, body string) error {
	return c.itdClient.Call(
		context.Background(),
		"Notify",
		NotifyData{
			Title: title,
			Body: body,
		},
		nil,
	)
}