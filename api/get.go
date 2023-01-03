package api

import (
	"context"

	"go.arsenm.dev/itd/internal/rpc"
)

func (c *Client) HeartRate(ctx context.Context) (uint8, error) {
	res, err := c.client.HeartRate(ctx, &rpc.Empty{})
	return uint8(res.Value), err
}

func (c *Client) BatteryLevel(ctx context.Context) (uint8, error) {
	res, err := c.client.BatteryLevel(ctx, &rpc.Empty{})
	return uint8(res.Value), err
}

type MotionValues struct {
	X, Y, Z int16
}

func (c *Client) Motion(ctx context.Context) (MotionValues, error) {
	res, err := c.client.Motion(ctx, &rpc.Empty{})
	return MotionValues{int16(res.X), int16(res.Y), int16(res.Z)}, err
}

func (c *Client) StepCount(ctx context.Context) (out uint32, err error) {
	res, err := c.client.StepCount(ctx, &rpc.Empty{})
	return res.Value, err
}

func (c *Client) Version(ctx context.Context) (out string, err error) {
	res, err := c.client.Version(ctx, &rpc.Empty{})
	return res.Value, err
}

func (c *Client) Address(ctx context.Context) (out string, err error) {
	res, err := c.client.Address(ctx, &rpc.Empty{})
	return res.Value, err
}
