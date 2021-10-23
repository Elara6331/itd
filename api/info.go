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
// new battery level values as they update
func (c *Client) WatchBatteryLevel() (<-chan uint8, error) {
	c.battLevelCh = make(chan uint8, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeBattLevel,
	})
	if err != nil {
		return nil, err
	}
	return c.battLevelCh, nil
}

// HeartRate gets the heart rate from the connected device
func (c *Client) HeartRate() (uint8, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeHeartRate,
	})
	if err != nil {
		return 0, err
	}

	return uint8(res.Value.(float64)), nil
}

// WatchHeartRate returns a channel which will contain
// new heart rate values as they update
func (c *Client) WatchHeartRate() (<-chan uint8, error) {
	c.heartRateCh = make(chan uint8, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchHeartRate,
	})
	if err != nil {
		return nil, err
	}
	return c.heartRateCh, nil
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
// new step count values as they update
func (c *Client) WatchStepCount() (<-chan uint32, error) {
	c.stepCountCh = make(chan uint32, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchStepCount,
	})
	if err != nil {
		return nil, err
	}
	return c.stepCountCh, nil
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
// new motion values as they update
func (c *Client) WatchMotion() (<-chan infinitime.MotionValues, error) {
	c.motionCh = make(chan infinitime.MotionValues, 2)
	err := c.requestNoRes(types.Request{
		Type: types.ReqTypeWatchMotion,
	})
	if err != nil {
		return nil, err
	}
	return c.motionCh, nil
}
