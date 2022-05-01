package api

func (c *Client) Remove(paths ...string) error {
	return c.client.Call(
		"FS",
		"Remove",
		paths,
		nil,
	)
}

func (c *Client) Rename(old, new string) error {
	return c.client.Call(
		"FS",
		"Rename",
		[2]string{old, new},
		nil,
	)
}

func (c *Client) Mkdir(paths ...string) error {
	return c.client.Call(
		"FS",
		"Mkdir",
		paths,
		nil,
	)
}

func (c *Client) ReadDir(dir string) (out []FileInfo, err error) {
	err = c.client.Call(
		"FS",
		"ReadDir",
		dir,
		&out,
	)
	return
}

func (c *Client) Upload(dst, src string) (chan FSTransferProgress, error) {
	progressCh := make(chan FSTransferProgress, 5)
	err := c.client.Call(
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

func (c *Client) Download(dst, src string) (chan FSTransferProgress, error) {
	progressCh := make(chan FSTransferProgress, 5)
	err := c.client.Call(
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
