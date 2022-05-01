package api

import (
	"go.arsenm.dev/infinitime"
)

func (c *Client) WatchHeartRate() (<-chan uint8, func(), error) {
	outCh := make(chan uint8, 2)
	err := c.client.Call(
		"ITD",
		"WatchHeartRate",
		nil,
		outCh,
	)
	if err != nil {
		return nil, nil, err
	}

	doneFn := func() {
		close(outCh)
	}

	return outCh, doneFn, nil
}

func (c *Client) WatchBatteryLevel() (<-chan uint8, func(), error) {
	outCh := make(chan uint8, 2)
	err := c.client.Call(
		"ITD",
		"WatchBatteryLevel",
		nil,
		outCh,
	)
	if err != nil {
		return nil, nil, err
	}

	doneFn := func() {
		close(outCh)
	}

	return outCh, doneFn, nil
}

func (c *Client) WatchStepCount() (<-chan uint32, func(), error) {
	outCh := make(chan uint32, 2)
	err := c.client.Call(
		"ITD",
		"WatchStepCount",
		nil,
		outCh,
	)
	if err != nil {
		return nil, nil, err
	}

	doneFn := func() {
		close(outCh)
	}

	return outCh, doneFn, nil
}

func (c *Client) WatchMotion() (<-chan infinitime.MotionValues, func(), error) {
	outCh := make(chan infinitime.MotionValues, 2)
	err := c.client.Call(
		"ITD",
		"WatchMotion",
		nil,
		outCh,
	)
	if err != nil {
		return nil, nil, err
	}

	doneFn := func() {
		close(outCh)
	}

	return outCh, doneFn, nil
}
