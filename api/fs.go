package api

import (
	"context"
	"errors"
	"io"

	"go.elara.ws/itd/internal/rpc"
)

type FSClient struct {
	client rpc.DRPCFSClient
}

func (c *FSClient) RemoveAll(ctx context.Context, paths ...string) error {
	_, err := c.client.RemoveAll(ctx, &rpc.PathsRequest{Paths: paths})
	return err
}

func (c *FSClient) Remove(ctx context.Context, paths ...string) error {
	_, err := c.client.Remove(ctx, &rpc.PathsRequest{Paths: paths})
	return err
}

func (c *FSClient) Rename(ctx context.Context, old, new string) error {
	_, err := c.client.Rename(ctx, &rpc.RenameRequest{
		From: old,
		To:   new,
	})
	return err
}

func (c *FSClient) MkdirAll(ctx context.Context, paths ...string) error {
	_, err := c.client.MkdirAll(ctx, &rpc.PathsRequest{Paths: paths})
	return err
}

func (c *FSClient) Mkdir(ctx context.Context, paths ...string) error {
	_, err := c.client.Mkdir(ctx, &rpc.PathsRequest{Paths: paths})
	return err
}

func (c *FSClient) ReadDir(ctx context.Context, dir string) ([]FileInfo, error) {
	res, err := c.client.ReadDir(ctx, &rpc.PathRequest{Path: dir})
	return convertEntries(res.Entries), err
}

func convertEntries(e []*rpc.FileInfo) []FileInfo {
	out := make([]FileInfo, len(e))
	for i, fi := range e {
		out[i] = FileInfo{
			Name:  fi.Name,
			Size:  fi.Size,
			IsDir: fi.IsDir,
		}
	}
	return out
}

func (c *FSClient) Upload(ctx context.Context, dst, src string) (chan FSTransferProgress, error) {
	progressCh := make(chan FSTransferProgress, 5)
	tc, err := c.client.Upload(ctx, &rpc.TransferRequest{Source: src, Destination: dst})
	if err != nil {
		return nil, err
	}

	go fsRecvToChannel[rpc.TransferProgress](tc, progressCh, func(evt *rpc.TransferProgress, err error) FSTransferProgress {
		return FSTransferProgress{
			Sent:  evt.Sent,
			Total: evt.Total,
			Err:   err,
		}
	})

	return progressCh, nil
}

func (c *FSClient) Download(ctx context.Context, dst, src string) (chan FSTransferProgress, error) {
	progressCh := make(chan FSTransferProgress, 5)
	tc, err := c.client.Download(ctx, &rpc.TransferRequest{Source: src, Destination: dst})
	if err != nil {
		return nil, err
	}

	go fsRecvToChannel[rpc.TransferProgress](tc, progressCh, func(evt *rpc.TransferProgress, err error) FSTransferProgress {
		return FSTransferProgress{
			Sent:  evt.Sent,
			Total: evt.Total,
			Err:   err,
		}
	})

	return progressCh, nil
}

// fsRecvToChannel converts a DRPC stream client to a Go channel, using cf to convert
// RPC generated types to API response types.
func fsRecvToChannel[R any, A any](s StreamClient[R], ch chan<- A, cf func(evt *R, err error) A) {
	defer close(ch)

	var err error
	var evt *R

	for {
		select {
		case <-s.Context().Done():
			return
		default:
			evt, err = s.Recv()
			if errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				ch <- cf(new(R), err)
				return
			}
			ch <- cf(evt, nil)
		}
	}
}
