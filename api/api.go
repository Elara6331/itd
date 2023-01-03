package api

import (
	"io"
	"net"

	"go.arsenm.dev/itd/internal/rpc"
	"storj.io/drpc/drpcconn"
)

const DefaultAddr = "/tmp/itd/socket"

type Client struct {
	conn   *drpcconn.Conn
	client rpc.DRPCITDClient
}

func New(sockPath string) (*Client, error) {
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, err
	}
	dconn := drpcconn.New(conn)

	return &Client{
		conn:   dconn,
		client: rpc.NewDRPCITDClient(dconn),
	}, nil
}

func NewFromConn(conn io.ReadWriteCloser) *Client {
	dconn := drpcconn.New(conn)

	return &Client{
		conn:   dconn,
		client: rpc.NewDRPCITDClient(dconn),
	}
}

func (c *Client) FS() *FSClient {
	return &FSClient{rpc.NewDRPCFSClient(c.conn)}
}

func (c *Client) Close() error {
	return c.conn.Close()
}
