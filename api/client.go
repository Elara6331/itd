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
	heartRateCh   chan types.Response
	battLevelCh   chan types.Response
	stepCountCh   chan types.Response
	motionCh      chan types.Response
	dfuProgressCh chan types.Response
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
		c.heartRateCh <- res
	case types.ResTypeWatchBattLevel:
		c.battLevelCh <- res
	case types.ResTypeWatchStepCount:
		c.stepCountCh <- res
	case types.ResTypeWatchMotion:
		c.motionCh <- res
	case types.ResTypeDFUProgress:
		c.dfuProgressCh <- res
	default:
		c.respCh <- res
	}
	return nil
}

func decodeUint8(val interface{}) uint8 {
	return uint8(val.(float64))
}

func decodeUint32(val interface{}) uint32 {
	return uint32(val.(float64))
}

func decodeMotion(val interface{}) (infinitime.MotionValues, error) {
	out := infinitime.MotionValues{}
	err := mapstructure.Decode(val, &out)
	if err != nil {
		return out, err
	}
	return out, nil
}

func decodeDFUProgress(val interface{}) (DFUProgress, error) {
	out := DFUProgress{}
	err := mapstructure.Decode(val, &out)
	if err != nil {
		return out, err
	}
	return out, nil
}
