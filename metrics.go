package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	"go.elara.ws/itd/infinitime"
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

	// Watch heart rate
	if k.Bool("metrics.heartRate.enabled") {
		err := dev.WatchHeartRate(ctx, func(heartRate uint8, err error) {
			if err != nil {
				// Handle error
				return
			}
			// Get current time
			unixTime := time.Now().UnixNano()
			// Insert sample and time into database
			db.Exec("INSERT INTO heartRate VALUES (?, ?);", unixTime, heartRate)
		})
		if err != nil {
			return err
		}
	}

	// If step count metrics enabled in config
	if k.Bool("metrics.stepCount.enabled") {
		// Watch step count
		err := dev.WatchStepCount(ctx, func(count uint32, err error) {
			if err != nil {
				return
			}
			// Get current time
			unixTime := time.Now().UnixNano()
			// Insert sample and time into database
			db.Exec("INSERT INTO stepCount VALUES (?, ?);", unixTime, count)
		})
		if err != nil {
			return err
		}
	}

	// Watch step count
	if k.Bool("metrics.stepCount.enabled") {
		err := dev.WatchStepCount(ctx, func(count uint32, err error) {
			if err != nil {
				// Handle error
				return
			}
			// Get current time
			unixTime := time.Now().UnixNano()
			// Insert sample and time into database
			db.Exec("INSERT INTO stepCount VALUES (?, ?);", unixTime, count)
		})
		if err != nil {
			return err
		}
	}

	// Watch battery level
	if k.Bool("metrics.battLevel.enabled") {
		err := dev.WatchBatteryLevel(ctx, func(battLevel uint8, err error) {
			if err != nil {
				// Handle error
				return
			}
			// Get current time
			unixTime := time.Now().UnixNano()
			// Insert sample and time into database
			db.Exec("INSERT INTO battLevel VALUES (?, ?);", unixTime, battLevel)
		})
		if err != nil {
			return err
		}
	}

	// Watch motion values
	if k.Bool("metrics.motion.enabled") {
		err := dev.WatchMotion(ctx, func(motionVals infinitime.MotionValues, err error) {
			if err != nil {
				// Handle error
				return
			}
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
		})
		if err != nil {
			return err
		}
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
