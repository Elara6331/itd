package api

import (
	"context"

	"go.arsenm.dev/infinitime"
)

func (c *Client) HeartRate() (out uint8, err error) {
	err = c.itdClient.Call(
		context.Background(),
		"HeartRate",
		nil,
		&out,
	)
	return
}

func (c *Client) BatteryLevel() (out uint8, err error) {
	err = c.itdClient.Call(
		context.Background(),
		"BatteryLevel",
		nil,
		&out,
	)
	return
}

func (c *Client) Motion() (out infinitime.MotionValues, err error) {
	err = c.itdClient.Call(
		context.Background(),
		"Motion",
		nil,
		&out,
	)
	return
}

func (c *Client) StepCount() (out uint32, err error) {
	err = c.itdClient.Call(
		context.Background(),
		"StepCount",
		nil,
		&out,
	)
	return
}

func (c *Client) Version() (out string, err error) {
	err = c.itdClient.Call(
		context.Background(),
		"Version",
		nil,
		&out,
	)
	return
}

func (c *Client) Address() (out string, err error) {
	err = c.itdClient.Call(
		context.Background(),
		"Address",
		nil,
		&out,
	)
	return
}
