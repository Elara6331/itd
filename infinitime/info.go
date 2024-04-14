package infinitime

import (
	"context"
	"encoding/binary"
)

// Address returns the MAC address of the connected device.
func (d *Device) Address() string {
	return d.device.Address.String()
}

// Version returns the version of InifniTime that the connected device is running.
func (d *Device) Version() (string, error) {
	c, err := d.getChar(firmwareVerChar)
	if err != nil {
		return "", err
	}

	ver := make([]byte, 16)
	n, err := c.Read(ver)
	return string(ver[:n]), err
}

// BatteryLevel returns the current battery level of the connected PineTime.
func (d *Device) BatteryLevel() (lvl uint8, err error) {
	c, err := d.getChar(batteryLevelChar)
	if err != nil {
		return 0, err
	}

	err = binary.Read(c, binary.LittleEndian, &lvl)
	return lvl, err
}

// WatchBatteryLevel calls fn whenever the battery level changes.
func (d *Device) WatchBatteryLevel(ctx context.Context, fn func(level uint8, err error)) error {
	return watchChar(ctx, d, batteryLevelChar, fn)
}

// StepCount returns the current step count recorded on the watch.
func (d *Device) StepCount() (sc uint32, err error) {
	c, err := d.getChar(stepCountChar)
	if err != nil {
		return 0, err
	}

	err = binary.Read(c, binary.LittleEndian, &sc)
	return sc, err
}

// WatchStepCount calls fn whenever the step count changes.
func (d *Device) WatchStepCount(ctx context.Context, fn func(count uint32, err error)) error {
	return watchChar(ctx, d, stepCountChar, fn)
}

// HeartRate returns the current heart rate recorded on the watch.
func (d *Device) HeartRate() (uint8, error) {
	c, err := d.getChar(heartRateChar)
	if err != nil {
		return 0, err
	}

	data := make([]byte, 2)
	_, err = c.Read(data)
	if err != nil {
		return 0, err
	}

	return data[1], nil
}

// WatchHeartRate calls fn whenever the heart rate changes.
func (d *Device) WatchHeartRate(ctx context.Context, fn func(rate uint8, err error)) error {
	return watchChar(ctx, d, heartRateChar, func(rate [2]uint8, err error) {
		fn(rate[1], err)
	})
}

// MotionValues represents gyroscope coordinates.
type MotionValues struct {
	X int16
	Y int16
	Z int16
}

// Motion returns the current gyroscope coordinates of the PineTime.
func (d *Device) Motion() (mv MotionValues, err error) {
	c, err := d.getChar(rawMotionChar)
	if err != nil {
		return MotionValues{}, err
	}

	err = binary.Read(c, binary.LittleEndian, &mv)
	return mv, err
}

// WatchMotion calls fn whenever the gyroscope coordinates change.
func (d *Device) WatchMotion(ctx context.Context, fn func(level MotionValues, err error)) error {
	return watchChar(ctx, d, rawMotionChar, fn)
}
