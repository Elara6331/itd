package api

import (
	"go.arsenm.dev/infinitime"
)

func (c *Client) HeartRate() (out uint8, err error) {
	err = c.client.Call(
		"ITD",
		"HeartRate",
		nil,
		&out,
	)
	return
}

func (c *Client) BatteryLevel() (out uint8, err error) {
	err = c.client.Call(
		"ITD",
		"BatteryLevel",
		nil,
		&out,
	)
	return
}

func (c *Client) Motion() (out infinitime.MotionValues, err error) {
	err = c.client.Call(
		"ITD",
		"Motion",
		nil,
		&out,
	)
	return
}

func (c *Client) StepCount() (out uint32, err error) {
	err = c.client.Call(
		"ITD",
		"StepCount",
		nil,
		&out,
	)
	return
}

func (c *Client) Version() (out string, err error) {
	err = c.client.Call(
		"ITD",
		"Version",
		nil,
		&out,
	)
	return
}

func (c *Client) Address() (out string, err error) {
	err = c.client.Call(
		"ITD",
		"Address",
		nil,
		&out,
	)
	return
}
