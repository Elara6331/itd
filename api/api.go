package api

import (
	"net"

	"go.arsenm.dev/lrpc/client"
	"go.arsenm.dev/lrpc/codec"
)

const DefaultAddr = "/tmp/itd/socket"

type Client struct {
	client *client.Client
}

func New(sockPath string) (*Client, error) {
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, err
	}

	out := &Client{
		client: client.New(conn, codec.JSON),
	}
	return out, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
