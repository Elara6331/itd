package api

import (
	"context"
	"time"
)

func (c *Client) Remove(paths ...string) error {
	return c.fsClient.Call(
		context.Background(),
		"Remove",
		paths,
		nil,
	)
}

func (c *Client) Rename(old, new string) error {
	return c.fsClient.Call(
		context.Background(),
		"Remove",
		[2]string{old, new},
		nil,
	)
}

func (c *Client) Mkdir(paths ...string) error {
	return c.fsClient.Call(
		context.Background(),
		"Mkdir",
		paths,
		nil,
	)
}

func (c *Client) ReadDir(dir string) (out []FileInfo, err error) {
	err = c.fsClient.Call(
		context.Background(),
		"ReadDir",
		dir,
		&out,
	)
	return
}

func (c *Client) Upload(dst, src string) (chan FSTransferProgress, error) {
	var id string
	err := c.fsClient.Call(
		context.Background(),
		"Upload",
		[2]string{dst, src},
		&id,
	)
	if err != nil {
		return nil, err
	}

	progressCh := make(chan FSTransferProgress, 5)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			progressCh <- val.(FSTransferProgress)
		}
		close(progressCh)
	}()

	return progressCh, nil
}

func (c *Client) Download(dst, src string) (chan FSTransferProgress, error) {
	var id string
	err := c.fsClient.Call(
		context.Background(),
		"Download",
		[2]string{dst, src},
		&id,
	)
	if err != nil {
		return nil, err
	}

	progressCh := make(chan FSTransferProgress, 5)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			progressCh <- val.(FSTransferProgress)
		}
		close(progressCh)
	}()

	return progressCh, nil
}
