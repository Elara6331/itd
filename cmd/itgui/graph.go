package main

import (
	"context"
	"database/sql"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/x/fyne/widget/charts"
	"go.elara.ws/itd/api"
	_ "modernc.org/sqlite"
)

func graphTab(ctx context.Context, client *api.Client, w fyne.Window) fyne.CanvasObject {
	// Get user configuration directory
	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}
	cfgDir := filepath.Join(userCfgDir, "itd")
	dbPath := filepath.Join(cfgDir, "metrics.db")

	// If stat on database returns error, return nil
	if _, err := os.Stat(dbPath); err != nil {
		return nil
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil
	}

	// Get heart rate data and create chart
	heartRateData := getData(db, "bpm", "heartRate")
	heartRate := newLineChartData(nil, heartRateData)

	// Get step count data and create chart
	stepCountData := getData(db, "steps", "stepCount")
	stepCount := newLineChartData(nil, stepCountData)

	// Get battery level data and create chart
	battLevelData := getData(db, "percent", "battLevel")
	battLevel := newLineChartData(nil, battLevelData)

	// Get motion data
	motionData := getMotionData(db)
	// Create chart for each coordinate
	xChart := newLineChartData(theme.PrimaryColorNamed(theme.ColorRed), motionData["X"])
	yChart := newLineChartData(theme.PrimaryColorNamed(theme.ColorGreen), motionData["Y"])
	zChart := newLineChartData(theme.PrimaryColorNamed(theme.ColorBlue), motionData["Z"])

	// Create new max container with all the charts
	motion := container.NewMax(xChart, yChart, zChart)

	// Create tabs for charts
	chartTabs := container.NewAppTabs(
		container.NewTabItem("Heart Rate", heartRate),
		container.NewTabItem("Step Count", stepCount),
		container.NewTabItem("Battery Level", battLevel),
		container.NewTabItem("Motion", motion),
	)
	// Place tabs on left
	chartTabs.SetTabLocation(container.TabLocationLeading)
	return chartTabs
}

func newLineChartData(col color.Color, data []float64) *charts.LineChart {
	// Create new line chart
	lc := charts.NewLineChart(nil)
	setOpts(lc, col)
	// If no data, make the stroke transparent
	if len(data) == 0 {
		lc.Options().StrokeColor = color.RGBA{0, 0, 0, 0}
	}
	// Set data
	lc.SetData(data)
	return lc
}

func setOpts(lc *charts.LineChart, col color.Color) {
	// Get pointer to options
	opts := lc.Options()
	// Set fill color to transparent
	opts.FillColor = color.RGBA{0, 0, 0, 0}
	// Set stroke width
	opts.StrokeWidth = 2
	// If color provided
	if col != nil {
		// Set stroke color
		opts.StrokeColor = col
	} else {
		// Set stroke color to orange primary color
		opts.StrokeColor = theme.PrimaryColorNamed(theme.ColorOrange)
	}
}

func getData(db *sql.DB, field, table string) []float64 {
	// Get data from database
	rows, err := db.Query("SELECT " + field + " FROM " + table + " ORDER BY time;")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []float64
	for rows.Next() {
		var val int64
		// Scan data into int
		err := rows.Scan(&val)
		if err != nil {
			return nil
		}

		// Convert to float64 and append to data slice
		out = append(out, float64(val))
	}

	return out
}

func getMotionData(db *sql.DB) map[string][]float64 {
	// Get data from database
	rows, err := db.Query("SELECT X, Y, Z FROM motion ORDER BY time;")
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := map[string][]float64{}
	for rows.Next() {
		var x, y, z int64
		// Scan data into ints
		err := rows.Scan(&x, &y, &z)
		if err != nil {
			return nil
		}

		// Convert to float64 and append to appropriate slice
		out["X"] = append(out["X"], float64(x))
		out["Y"] = append(out["Y"], float64(y))
		out["Z"] = append(out["Z"], float64(z))
	}

	return out
}
