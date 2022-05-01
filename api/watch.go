package api

import (
	"context"

	"go.arsenm.dev/infinitime"
)

func (c *Client) WatchHeartRate(ctx context.Context) (<-chan uint8, error) {
	outCh := make(chan uint8, 2)
	err := c.client.Call(
		ctx,
		"ITD",
		"WatchHeartRate",
		nil,
		outCh,
	)
	if err != nil {
		return nil, err
	}

	return outCh, nil
}

func (c *Client) WatchBatteryLevel(ctx context.Context) (<-chan uint8, error) {
	outCh := make(chan uint8, 2)
	err := c.client.Call(
		ctx,
		"ITD",
		"WatchBatteryLevel",
		nil,
		outCh,
	)
	if err != nil {
		return nil, err
	}

	return outCh, nil
}

func (c *Client) WatchStepCount(ctx context.Context) (<-chan uint32, error) {
	outCh := make(chan uint32, 2)
	err := c.client.Call(
		ctx,
		"ITD",
		"WatchStepCount",
		nil,
		outCh,
	)
	if err != nil {
		return nil, err
	}

	return outCh, nil
}

func (c *Client) WatchMotion(ctx context.Context) (<-chan infinitime.MotionValues, error) {
	outCh := make(chan infinitime.MotionValues, 2)
	err := c.client.Call(
		ctx,
		"ITD",
		"WatchMotion",
		nil,
		outCh,
	)
	if err != nil {
		return nil, err
	}

	return outCh, nil
}
