package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	"go.elara.ws/infinitime"
	"go.elara.ws/logger/log"
	_ "modernc.org/sqlite"
)

func initMetrics(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	// If metrics disabled, return nil
	if !k.Bool("metrics.enabled") {
		return nil
	}

	// Open metrics database
	db, err := sql.Open("sqlite", filepath.Join(cfgDir, "metrics.db"))
	if err != nil {
		return err
	}

	// Create heartRate table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS heartRate(time INT, bpm INT);")
	if err != nil {
		return err
	}

	// Create stepCount table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS stepCount(time INT, steps INT);")
	if err != nil {
		return err
	}

	// Create battLevel table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS battLevel(time INT, percent INT);")
	if err != nil {
		return err
	}

	// Create motion table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS motion(time INT, X INT, Y INT, Z INT);")
	if err != nil {
		return err
	}

	// If heart rate metrics enabled in config
	if k.Bool("metrics.heartRate.enabled") {
		// Watch heart rate
		heartRateCh, err := dev.WatchHeartRate(ctx)
		if err != nil {
			return err
		}
		go func() {
			// For every heart rate sample
			for heartRate := range heartRateCh {
				// Get current time
				unixTime := time.Now().UnixNano()
				// Insert sample and time into database
				db.Exec("INSERT INTO heartRate VALUES (?, ?);", unixTime, heartRate)
			}
		}()
	}

	// If step count metrics enabled in config
	if k.Bool("metrics.stepCount.enabled") {
		// Watch step count
		stepCountCh, err := dev.WatchStepCount(ctx)
		if err != nil {
			return err
		}
		go func() {
			// For every step count sample
			for stepCount := range stepCountCh {
				// Get current time
				unixTime := time.Now().UnixNano()
				// Insert sample and time into database
				db.Exec("INSERT INTO stepCount VALUES (?, ?);", unixTime, stepCount)
			}
		}()
	}

	// If battery level metrics enabled in config
	if k.Bool("metrics.battLevel.enabled") {
		// Watch battery level
		battLevelCh, err := dev.WatchBatteryLevel(ctx)
		if err != nil {
			return err
		}
		go func() {
			// For every battery level sample
			for battLevel := range battLevelCh {
				// Get current time
				unixTime := time.Now().UnixNano()
				// Insert sample and time into database
				db.Exec("INSERT INTO battLevel VALUES (?, ?);", unixTime, battLevel)
			}
		}()
	}

	// If motion metrics enabled in config
	if k.Bool("metrics.motion.enabled") {
		// Watch motion values
		motionCh, err := dev.WatchMotion(ctx)
		if err != nil {
			return err
		}
		go func() {
			// For every motion sample
			for motionVals := range motionCh {
				// Get current time
				unixTime := time.Now().UnixNano()
				// Insert sample values and time into database
				db.Exec(
					"INSERT INTO motion VALUES (?, ?, ?, ?);",
					unixTime,
					motionVals.X,
					motionVals.Y,
					motionVals.Z,
				)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done("metrics")
		<-ctx.Done()
		db.Close()
	}()

	log.Info("Initialized metrics collection").Send()

	return nil
}
