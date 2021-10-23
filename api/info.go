package api

import (
	"github.com/mitchellh/mapstructure"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/itd/internal/types"
)

func (c *Client) Address() (string, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeBtAddress,
	})
	if err != nil {
		return "", err
	}

	return res.Value.(string), nil
}

func (c *Client) Version() (string, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeFwVersion,
	})
	if err != nil {
		return "", err
	}

	return res.Value.(string), nil
}

func (c *Client) BatteryLevel() (uint8, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeBattLevel,
	})
	if err != nil {
		return 0, err
	}

	return uint8(res.Value.(float64)), nil
}

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

func (c *Client) HeartRate() (uint8, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeHeartRate,
	})
	if err != nil {
		return 0, err
	}

	return uint8(res.Value.(float64)), nil
}

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

func (c *Client) StepCount() (uint32, error) {
	res, err := c.request(types.Request{
		Type: types.ReqTypeStepCount,
	})
	if err != nil {
		return 0, err
	}

	return uint32(res.Value.(float64)), nil
}

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
