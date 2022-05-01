package api

func (c *Client) Notify(title, body string) error {
	return c.client.Call(
		"ITD",
		"Notify",
		NotifyData{
			Title: title,
			Body:  body,
		},
		nil,
	)
}
