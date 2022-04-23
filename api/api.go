package api

import (
	"context"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/vmihailenco/msgpack/v5"
	"go.arsenm.dev/infinitime"
)

const DefaultAddr = "/tmp/itd/socket"

type Client struct {
	itdClient client.XClient
	itdCh     chan *protocol.Message
	fsClient  client.XClient
	fsCh      chan *protocol.Message
	srvVals   map[string]chan interface{}
}

func New(sockPath string) (*Client, error) {
	d, err := client.NewPeer2PeerDiscovery("unix@"+sockPath, "")
	if err != nil {
		return nil, err
	}

	out := &Client{}

	out.itdCh = make(chan *protocol.Message, 5)
	out.itdClient = client.NewBidirectionalXClient(
		"ITD",
		client.Failtry,
		client.RandomSelect,
		d,
		client.DefaultOption,
		out.itdCh,
	)

	out.fsCh = make(chan *protocol.Message, 5)
	out.fsClient = client.NewBidirectionalXClient(
		"FS",
		client.Failtry,
		client.RandomSelect,
		d,
		client.DefaultOption,
		out.fsCh,
	)

	out.srvVals = map[string]chan interface{}{}

	go out.handleMessages(out.itdCh)
	go out.handleMessages(out.fsCh)

	return out, nil
}

func (c *Client) handleMessages(msgCh chan *protocol.Message) {
	for msg := range msgCh {
		_, ok := c.srvVals[msg.ServicePath]
		if !ok {
			c.srvVals[msg.ServicePath] = make(chan interface{}, 5)
		}

		//fmt.Printf("%+v\n", msg)

		ch := c.srvVals[msg.ServicePath]

		switch msg.ServiceMethod {
		case "FSProgress":
			var progress FSTransferProgress
			msgpack.Unmarshal(msg.Payload, &progress)
			ch <- progress
		case "DFUProgress":
			var progress infinitime.DFUProgress
			msgpack.Unmarshal(msg.Payload, &progress)
			ch <- progress
		case "MotionSample":
			var motionVals infinitime.MotionValues
			msgpack.Unmarshal(msg.Payload, &motionVals)
			ch <- motionVals
		case "Done":
			close(c.srvVals[msg.ServicePath])
			delete(c.srvVals, msg.ServicePath)
		default:
			var value interface{}
			msgpack.Unmarshal(msg.Payload, &value)
			ch <- value
		}
	}
}

func (c *Client) done(id string) error {
	return c.itdClient.Call(
		context.Background(),
		"Done",
		id,
		nil,
	)
}

func (c *Client) Close() error {
	err := c.itdClient.Close()
	if err != nil {
		return err
	}

	err = c.fsClient.Close()
	if err != nil {
		return err
	}

	close(c.itdCh)
	close(c.fsCh)

	return nil
}
