package api

import (
	"io"
	"net"

	"go.arsenm.dev/itd/internal/rpc"
	"storj.io/drpc"
)

const DefaultAddr = "/tmp/itd/socket"

type Client struct {
	conn   drpc.Conn
	client rpc.DRPCITDClient
}

func New(sockPath string) (*Client, error) {
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, err
	}

	mconn, err := newMuxConn(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   mconn,
		client: rpc.NewDRPCITDClient(mconn),
	}, nil
}

func NewFromConn(conn io.ReadWriteCloser) (*Client, error) {
	mconn, err := newMuxConn(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   mconn,
		client: rpc.NewDRPCITDClient(mconn),
	}, nil
}

func (c *Client) FS() *FSClient {
	return &FSClient{rpc.NewDRPCFSClient(c.conn)}
}

func (c *Client) Close() error {
	return c.conn.Close()
}
