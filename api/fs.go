package api

import "context"

func (c *Client) RemoveAll(ctx context.Context, paths ...string) error {
	return c.client.Call(
		ctx,
		"FS",
		"RemoveAll",
		paths,
		nil,
	)
}

func (c *Client) Remove(ctx context.Context, paths ...string) error {
	return c.client.Call(
		ctx,
		"FS",
		"Remove",
		paths,
		nil,
	)
}

func (c *Client) Rename(ctx context.Context, old, new string) error {
	return c.client.Call(
		ctx,
		"FS",
		"Rename",
		[2]string{old, new},
		nil,
	)
}

func (c *Client) MkdirAll(ctx context.Context, paths ...string) error {
	return c.client.Call(
		ctx,
		"FS",
		"MkdirAll",
		paths,
		nil,
	)
}

func (c *Client) Mkdir(ctx context.Context, paths ...string) error {
	return c.client.Call(
		ctx,
		"FS",
		"Mkdir",
		paths,
		nil,
	)
}

func (c *Client) ReadDir(ctx context.Context, dir string) (out []FileInfo, err error) {
	err = c.client.Call(
		ctx,
		"FS",
		"ReadDir",
		dir,
		&out,
	)
	return
}

func (c *Client) Upload(ctx context.Context, dst, src string) (chan FSTransferProgress, error) {
	progressCh := make(chan FSTransferProgress, 5)
	err := c.client.Call(
		ctx,
		"FS",
		"Upload",
		[2]string{dst, src},
		progressCh,
	)
	if err != nil {
		return nil, err
	}

	return progressCh, nil
}

func (c *Client) Download(ctx context.Context, dst, src string) (chan FSTransferProgress, error) {
	progressCh := make(chan FSTransferProgress, 5)
	err := c.client.Call(
		ctx,
		"FS",
		"Download",
		[2]string{dst, src},
		progressCh,
	)
	if err != nil {
		return nil, err
	}

	return progressCh, nil
}
