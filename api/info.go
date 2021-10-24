package api

import (
	"github.com/mitchellh/mapstructure"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/types"
)

// Address gets the bluetooth address of the connected device
func (c *Client) Address() (string, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeBtAddress,
	})
	if err != nil {
		return "", err
	}

	return res.Value.(string), nil
}

// Version gets the firmware version of the connected device
func (c *Client) Version() (string, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeFwVersion,
	})
	if err != nil {
		return "", err
	}

	return res.Value.(string), nil
}

// BatteryLevel gets the battery level of the connected device
func (c *Client) BatteryLevel() (uint8, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeBattLevel,
	})
	if err != nil {
		return 0, err
	}

	return uint8(res.Value.(float64)), nil
}

// WatchBatteryLevel returns a channel which will contain
// new battery level values as they update. Do not use after
// calling cancellation function
func (c *Client) WatchBatteryLevel() (<-chan uint8, func(), error) {
	c.battLevelCh = make(chan types.Response, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchBattLevel,
	})
	if err != nil {
		return nil, nil, err
	}
	res := <-c.battLevelCh
	done, cancel := c.cancelFn(res.ID, c.battLevelCh)
	out := make(chan uint8, 2)
	go func() {
		for res := range c.battLevelCh {
			select {
			case <-done:
				return
			default:
				out <- decodeUint8(res.Value)
			}
		}
	}()
	return out, cancel, nil
}

// HeartRate gets the heart rate from the connected device
func (c *Client) HeartRate() (uint8, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeHeartRate,
	})
	if err != nil {
		return 0, err
	}

	return decodeUint8(res.Value), nil
}

// WatchHeartRate returns a channel which will contain
// new heart rate values as they update. Do not use after
// calling cancellation function
func (c *Client) WatchHeartRate() (<-chan uint8, func(), error) {
	c.heartRateCh = make(chan types.Response, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchHeartRate,
	})
	if err != nil {
		return nil, nil, err
	}
	res := <-c.heartRateCh
	done, cancel := c.cancelFn(res.ID, c.heartRateCh)
	out := make(chan uint8, 2)
	go func() {
		for res := range c.heartRateCh {
			select {
			case <-done:
				return
			default:
				out <- decodeUint8(res.Value)
			}
		}
	}()
	return out, cancel, nil
}

// cancelFn generates a cancellation function for the given
// request type and channel
func (c *Client) cancelFn(reqID string, ch chan types.Response) (chan struct{}, func()) {
	done := make(chan struct{}, 1)
	return done, func() {
		done <- struct{}{}
		close(ch)
		c.requestNoRes(types.Request{
			Type: types.ReqTypeCancel,
			Data: reqID,
		})
	}
}

// StepCount gets the step count from the connected device
func (c *Client) StepCount() (uint32, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeStepCount,
	})
	if err != nil {
		return 0, err
	}

	return uint32(res.Value.(float64)), nil
}

// WatchStepCount returns a channel which will contain
// new step count values as they update. Do not use after
// calling cancellation function
func (c *Client) WatchStepCount() (<-chan uint32, func(), error) {
	c.stepCountCh = make(chan types.Response, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchStepCount,
	})
	if err != nil {
		return nil, nil, err
	}
	res := <-c.stepCountCh
	done, cancel := c.cancelFn(res.ID, c.stepCountCh)
	out := make(chan uint32, 2)
	go func() {
		for res := range c.stepCountCh {
			select {
			case <-done:
				return
			default:
				out <- decodeUint32(res.Value)
			}
		}
	}()
	return out, cancel, nil
}

// Motion gets the motion values from the connected device
func (c *Client) Motion() (infinitime.MotionValues, error) {
	out := infinitime.MotionValues{}
	res, err := c.request(types.Request{
		Type: types.ReqTypeMotion,
	})
	if err != nil {
		return out, err
	}
	err = mapstructure.Decode(res.Value, &out)
	if err != nil {
		return out, err
	}
	return out, nil
}

// WatchMotion returns a channel which will contain
// new motion values as they update. Do not use after
// calling cancellation function
func (c *Client) WatchMotion() (<-chan infinitime.MotionValues, func(), error) {
	c.motionCh = make(chan types.Response, 5)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchMotion,
	})
	if err != nil {
		return nil, nil, err
	}
	res := <-c.motionCh
	done, cancel := c.cancelFn(res.ID, c.motionCh)
	out := make(chan infinitime.MotionValues, 5)
	go func() {
		for res := range c.motionCh {
			select {
			case <-done:
				return
			default:
				motion, err := decodeMotion(res.Value)
				if err != nil {
					continue
				}
				out <- motion
			}
		}
	}()
	return out, cancel, nil
}
