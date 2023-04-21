package api

import (
	"context"

	"go.elara.ws/itd/internal/rpc"
)

func (c *Client) WatchHeartRate(ctx context.Context) (<-chan uint8, error) {
	outCh := make(chan uint8, 2)
	wc, err := c.client.WatchHeartRate(ctx, &rpc.Empty{})
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outCh)

		var err error
		var evt *rpc.IntResponse

		for {
			select {
			case <-ctx.Done():
				wc.Close()
				return
			default:
				evt, err = wc.Recv()
				if err != nil {
					return
				}
			}

			outCh <- uint8(evt.Value)
		}
	}()

	return outCh, nil
}

func (c *Client) WatchBatteryLevel(ctx context.Context) (<-chan uint8, error) {
	outCh := make(chan uint8, 2)
	wc, err := c.client.WatchBatteryLevel(ctx, &rpc.Empty{})
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outCh)

		var err error
		var evt *rpc.IntResponse

		for {
			select {
			case <-ctx.Done():
				wc.Close()
				return
			default:
				evt, err = wc.Recv()
				if err != nil {
					return
				}
			}

			outCh <- uint8(evt.Value)
		}
	}()

	return outCh, nil
}

func (c *Client) WatchStepCount(ctx context.Context) (<-chan uint32, error) {
	outCh := make(chan uint32, 2)
	wc, err := c.client.WatchStepCount(ctx, &rpc.Empty{})
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outCh)

		var err error
		var evt *rpc.IntResponse

		for {
			select {
			case <-ctx.Done():
				wc.Close()
				return
			default:
				evt, err = wc.Recv()
				if err != nil {
					return
				}
			}

			outCh <- evt.Value
		}
	}()

	return outCh, nil
}

func (c *Client) WatchMotion(ctx context.Context) (<-chan MotionValues, error) {
	outCh := make(chan MotionValues, 2)
	wc, err := c.client.WatchMotion(ctx, &rpc.Empty{})
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outCh)

		var err error
		var evt *rpc.MotionResponse

		for {
			select {
			case <-ctx.Done():
				wc.Close()
				return
			default:
				evt, err = wc.Recv()
				if err != nil {
					return
				}
			}

			outCh <- MotionValues{int16(evt.X), int16(evt.Y), int16(evt.Z)}
		}
	}()

	return outCh, nil
}
