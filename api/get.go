package api

import (
	"context"

	"go.arsenm.dev/infinitime"
)

func (c *Client) HeartRate(ctx context.Context) (out uint8, err error) {
	err = c.client.Call(
		ctx,
		"ITD",
		"HeartRate",
		nil,
		&out,
	)
	return
}

func (c *Client) BatteryLevel(ctx context.Context) (out uint8, err error) {
	err = c.client.Call(
		ctx,
		"ITD",
		"BatteryLevel",
		nil,
		&out,
	)
	return
}

func (c *Client) Motion(ctx context.Context) (out infinitime.MotionValues, err error) {
	err = c.client.Call(
		ctx,
		"ITD",
		"Motion",
		nil,
		&out,
	)
	return
}

func (c *Client) StepCount(ctx context.Context) (out uint32, err error) {
	err = c.client.Call(
		ctx,
		"ITD",
		"StepCount",
		nil,
		&out,
	)
	return
}

func (c *Client) Version(ctx context.Context) (out string, err error) {
	err = c.client.Call(
		ctx,
		"ITD",
		"Version",
		nil,
		&out,
	)
	return
}

func (c *Client) Address(ctx context.Context) (out string, err error) {
	err = c.client.Call(
		ctx,
		"ITD",
		"Address",
		nil,
		&out,
	)
	return
}
