package api

import (
	"context"

	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/rpc"
)

type ResourceOperation uint8

const (
	ResourceOperationRemoveObsolete = infinitime.ResourceOperationRemoveObsolete
	ResourceOperationUpload         = infinitime.ResourceOperationUpload
)

type ResourceLoadProgress struct {
	Operation ResourceOperation
	Name      string
	Total     int64
	Sent      int64
	Err       error
}

// LoadResources loads resources onto the watch from the given
// file path to the resources zip
func (c *FSClient) LoadResources(ctx context.Context, path string) (<-chan ResourceLoadProgress, error) {
	progCh := make(chan ResourceLoadProgress, 2)

	rc, err := c.client.LoadResources(ctx, &rpc.PathRequest{Path: path})
	if err != nil {
		return nil, err
	}

	go fsRecvToChannel[rpc.ResourceLoadProgress](rc, progCh, func(evt *rpc.ResourceLoadProgress, err error) ResourceLoadProgress {
		return ResourceLoadProgress{
			Operation: ResourceOperation(evt.Operation),
			Name:      evt.Name,
			Sent:      evt.Sent,
			Total:     evt.Total,
			Err:       err,
		}
	})

	return progCh, nil
}

type StreamClient[T any] interface {
	Recv() (*T, error)
	Context() context.Context
}
