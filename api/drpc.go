package api

import (
	"context"
	"io"

	"github.com/hashicorp/yamux"
	"storj.io/drpc"
	"storj.io/drpc/drpcconn"
)

var _ drpc.Conn = &muxConn{}

type muxConn struct {
	conn   io.ReadWriteCloser
	sess   *yamux.Session
	closed chan struct{}
}

func newMuxConn(conn io.ReadWriteCloser) (*muxConn, error) {
	sess, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	return &muxConn{
		conn:   conn,
		sess:   sess,
		closed: make(chan struct{}),
	}, nil
}

func (m *muxConn) Close() error {
	defer close(m.closed)

	err := m.sess.Close()
	if err != nil {
		return err
	}
	return m.conn.Close()
}

func (m *muxConn) Closed() <-chan struct{} {
	return m.closed
}

func (m *muxConn) Invoke(ctx context.Context, rpc string, enc drpc.Encoding, in, out drpc.Message) error {
	conn, err := m.sess.Open()
	if err != nil {
		return err
	}
	defer conn.Close()
	dconn := drpcconn.New(conn)
	return dconn.Invoke(ctx, rpc, enc, in, out)
}

func (m *muxConn) NewStream(ctx context.Context, rpc string, enc drpc.Encoding) (drpc.Stream, error) {
	conn, err := m.sess.Open()
	if err != nil {
		return nil, err
	}

	dconn := drpcconn.New(conn)
	return dconn.NewStream(ctx, rpc, enc)
}