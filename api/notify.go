package api

import "context"

func (c *Client) Notify(ctx context.Context, title, body string) error {
	return c.client.Call(
		ctx,
		"ITD",
		"Notify",
		NotifyData{
			Title: title,
			Body:  body,
		},
		nil,
	)
}
