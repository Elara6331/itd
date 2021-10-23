package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"

	"github.com/mitchellh/mapstructure"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/types"
)

// Default socket address
const DefaultAddr = "/tmp/itd/socket"

// Client is the socket API client
type Client struct {
	conn          net.Conn
	respCh        chan types.Response
	heartRateCh   chan uint8
	battLevelCh   chan uint8
	stepCountCh   chan uint32
	motionCh      chan infinitime.MotionValues
	dfuProgressCh chan DFUProgress
}

// New creates a new client and sets it up
func New(addr string) (*Client, error) {
	conn, err := net.Dial("unix", addr)
	if err != nil {
		return nil, err
	}

	out := &Client{
		conn:   conn,
		respCh: make(chan types.Response, 5),
	}

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			var res types.Response
			err = json.Unmarshal(scanner.Bytes(), &res)
			if err != nil {
				continue
			}
			out.handleResp(res)
		}
	}()
	return out, err
}

func (c *Client) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	close(c.respCh)
	return nil
}

// request sends a request to itd and waits for and returns the response
func (c *Client) request(req types.Request) (types.Response, error) {
	// Encode request into connection
	err := json.NewEncoder(c.conn).Encode(req)
	if err != nil {
		return types.Response{}, err
	}

	res := <-c.respCh

	if res.Error {
		return res, errors.New(res.Message)
	}

	return res, nil
}

// requestNoRes sends a request to itd and does not wait for the response
func (c *Client) requestNoRes(req types.Request) error {
	// Encode request into connection
	err := json.NewEncoder(c.conn).Encode(req)
	if err != nil {
		return err
	}
	return nil
}

// handleResp handles the received response as needed
func (c *Client) handleResp(res types.Response) error {
	switch res.Type {
	case types.ResTypeWatchHeartRate:
		c.heartRateCh <- uint8(res.Value.(float64))
	case types.ResTypeWatchBattLevel:
		c.battLevelCh <- uint8(res.Value.(float64))
	case types.ResTypeWatchStepCount:
		c.stepCountCh <- uint32(res.Value.(float64))
	case types.ResTypeWatchMotion:
		out := infinitime.MotionValues{}
		err := mapstructure.Decode(res.Value, &out)
		if err != nil {
			return err
		}
		c.motionCh <- out
	case types.ResTypeDFUProgress:
		out := DFUProgress{}
		err := mapstructure.Decode(res.Value, &out)
		if err != nil {
			return err
		}
		c.dfuProgressCh <- out
	default:
		c.respCh <- res
	}
	return nil
}
