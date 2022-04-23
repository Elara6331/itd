package api

import (
	"context"
	"time"

	"go.arsenm.dev/infinitime"
)

func (c *Client) WatchHeartRate() (<-chan uint8, func(), error) {
	var id string
	err := c.itdClient.Call(
		context.Background(),
		"WatchHeartRate",
		nil,
		&id,
	)
	if err != nil {
		return nil, nil, err
	}

	outCh := make(chan uint8, 2)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			outCh <- val.(uint8)
		}
	}()

	doneFn := func() {
		c.done(id)
		close(c.srvVals[id])
		delete(c.srvVals, id)
	}

	return outCh, doneFn, nil
}

func (c *Client) WatchBatteryLevel() (<-chan uint8, func(), error) {
	var id string
	err := c.itdClient.Call(
		context.Background(),
		"WatchBatteryLevel",
		nil,
		&id,
	)
	if err != nil {
		return nil, nil, err
	}

	outCh := make(chan uint8, 2)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			outCh <- val.(uint8)
		}
	}()

	doneFn := func() {
		c.done(id)
		close(c.srvVals[id])
		delete(c.srvVals, id)
	}

	return outCh, doneFn, nil
}

func (c *Client) WatchStepCount() (<-chan uint32, func(), error) {
	var id string
	err := c.itdClient.Call(
		context.Background(),
		"WatchStepCount",
		nil,
		&id,
	)
	if err != nil {
		return nil, nil, err
	}

	outCh := make(chan uint32, 2)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			outCh <- val.(uint32)
		}
	}()

	doneFn := func() {
		c.done(id)
		close(c.srvVals[id])
		delete(c.srvVals, id)
	}

	return outCh, doneFn, nil
}

func (c *Client) WatchMotion() (<-chan infinitime.MotionValues, func(), error) {
	var id string
	err := c.itdClient.Call(
		context.Background(),
		"WatchMotion",
		nil,
		&id,
	)
	if err != nil {
		return nil, nil, err
	}

	outCh := make(chan infinitime.MotionValues, 2)
	go func() {
		srvValCh, ok := c.srvVals[id]
		for !ok {
			time.Sleep(100 * time.Millisecond)
			srvValCh, ok = c.srvVals[id]
		}

		for val := range srvValCh {
			outCh <- val.(infinitime.MotionValues)
		}
	}()

	doneFn := func() {
		c.done(id)
		close(c.srvVals[id])
		delete(c.srvVals, id)
	}

	return outCh, doneFn, nil
}
