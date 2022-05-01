package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func getAddress(c *cli.Context) error {
	address, err := client.Address(c.Context)
	if err != nil {
		return err
	}

	fmt.Println(address)
	return nil
}

func getBattery(c *cli.Context) error {
	battLevel, err := client.BatteryLevel(c.Context)
	if err != nil {
		return err
	}

	// Print returned percentage
	fmt.Printf("%d%%\n", battLevel)
	return nil
}

func getHeart(c *cli.Context) error {
	heartRate, err := client.HeartRate(c.Context)
	if err != nil {
		return err
	}

	// Print returned BPM
	fmt.Printf("%d BPM\n", heartRate)
	return nil
}

func getMotion(c *cli.Context) error {
	motionVals, err := client.Motion(c.Context)
	if err != nil {
		return err
	}

	if c.Bool("shell") {
		fmt.Printf(
			"X=%d\nY=%d\nZ=%d\n",
			motionVals.X,
			motionVals.Y,
			motionVals.Z,
		)
	} else {
		return json.NewEncoder(os.Stdout).Encode(motionVals)
	}
	return nil
}

func getSteps(c *cli.Context) error {
	stepCount, err := client.StepCount(c.Context)
	if err != nil {
		return err
	}

	// Print returned BPM
	fmt.Printf("%d Steps\n", stepCount)
	return nil
}
