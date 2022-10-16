package api

import (
	"context"

	"go.arsenm.dev/infinitime"
)

// LoadResources loads resources onto the watch from the given
// file path to the resources zip
func (c *Client) LoadResources(ctx context.Context, path string) (<-chan infinitime.ResourceLoadProgress, error) {
	progCh := make(chan infinitime.ResourceLoadProgress)

	err := c.client.Call(
		ctx,
		"FS",
		"LoadResources",
		path,
		progCh,
	)
	if err != nil {
		return nil, err
	}

	return progCh, nil
}
