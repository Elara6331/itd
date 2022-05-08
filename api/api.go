package api

import (
	"io"
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
		client: client.New(conn, codec.Default),
	}
	return out, nil
}

func NewFromConn(conn io.ReadWriteCloser) *Client {
	return &Client{
		client: client.New(conn, codec.Default),
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}
