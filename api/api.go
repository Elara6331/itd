package api

import (
	"io"
	"net"

	"go.elara.ws/drpc/muxconn"
	"go.elara.ws/itd/internal/rpc"
	"storj.io/drpc"
)

const DefaultAddr = "/tmp/itd/socket"

// Client is a client for ITD's socket API
type Client struct {
	conn   drpc.Conn
	client rpc.DRPCITDClient
}

// New connects to the UNIX socket at the given
// path, and returns a client that communicates
// with that socket.
func New(sockPath string) (*Client, error) {
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, err
	}

	mconn, err := muxconn.New(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   mconn,
		client: rpc.NewDRPCITDClient(mconn),
	}, nil
}

// NewFromConn returns a client that communicates
// over the given connection.
func NewFromConn(conn io.ReadWriteCloser) (*Client, error) {
	mconn, err := muxconn.New(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   mconn,
		client: rpc.NewDRPCITDClient(mconn),
	}, nil
}

// FS returns the filesystem API client
func (c *Client) FS() *FSClient {
	return &FSClient{rpc.NewDRPCFSClient(c.conn)}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}
