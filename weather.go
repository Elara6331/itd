package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.elara.ws/itd/infinitime"
	"go.elara.ws/logger/log"
)

// METResponse represents a response from
// the MET Norway API
type METResponse struct {
	Properties struct {
		Timeseries []struct {
			Time time.Time
			Data METData
		}
	}
}

// METData represents data in a METResponse
type METData struct {
	Instant struct {
		Details struct {
			AirPressure       float32 `json:"air_pressure_at_sea_level"`
			Temperature       float32 `json:"air_temperature"`
			DewPoint          float32 `json:"dew_point_temperature"`
			CloudAreaFraction float32 `json:"cloud_area_fraction"`
			FogAreaFraction   float32 `json:"fog_area_fraction"`
			RelativeHumidity  float32 `json:"relative_humidity"`
			UVIndex           float32 `json:"ultraviolet_index_clear_sky"`
			WindDirection     float32 `json:"wind_from_direction"`
			WindSpeed         float32 `json:"wind_speed"`
		}
	}
	NextHour struct {
		Summary struct {
			SymbolCode string `json:"symbol_code"`
		}
		Details struct {
			PrecipitationAmount float32 `json:"precipitation_amount"`
		}
	} `json:"next_1_hours"`
	Next6Hours struct {
		Details struct {
			MaxTemp float32 `json:"air_temperature_max"`
			MinTemp float32 `json:"air_temperature_min"`
		}
	} `json:"next_6_hours"`
}

// OSMData represents lat/long data from
// OpenStreetMap Nominatim
type OSMData []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

var sendWeatherCh = make(chan struct{}, 1)

func sleepCtx(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

func initWeather(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	if !k.Bool("weather.enabled") {
		return nil
	}

	// Get location based on string in config
	lat, lon, err := getLocation(ctx, k.String("weather.location"))
	if err != nil {
		return err
	}

	timer := time.NewTimer(time.Hour)

	wg.Add(1)
	go func() {
		defer wg.Done("weather")
		for {
			select {
			case _, ok := <-ctx.Done():
				if !ok {
					return
				}
			default:
			}

			// Attempt to get weather
			data, err := getWeather(ctx, lat, lon)
			if err != nil {
				log.Warn("Error getting weather data").Err(err).Send()
				// Wait 15 minutes before retrying
				sleepCtx(ctx, 15*time.Minute)
				continue
			}

			// Get current data
			current := data.Properties.Timeseries[0]
			currentData := current.Data.Instant.Details

			icon := parseSymbol(current.Data.NextHour.Summary.SymbolCode)
			if icon == infinitime.WeatherIconClear {
				switch {
				case currentData.CloudAreaFraction > 0.5:
					icon = infinitime.WeatherIconHeavyClouds
				case currentData.CloudAreaFraction == 0.5:
					icon = infinitime.WeatherIconClouds
				case currentData.CloudAreaFraction > 0:
					icon = infinitime.WeatherIconFewClouds
				}
			}

			err = dev.SetCurrentWeather(infinitime.CurrentWeather{
				Time:        time.Now(),
				CurrentTemp: currentData.Temperature,
				MaxTemp:     current.Data.Next6Hours.Details.MaxTemp,
				MinTemp:     current.Data.Next6Hours.Details.MinTemp,
				Location:    k.String("weather.location"),
				Icon:        icon,
			})
			if err != nil {
				log.Error("Error setting weather").Err(err).Send()
			}

			// Reset timer to 1 hour
			timer.Stop()
			timer.Reset(time.Hour)

			// Wait for timer to fire or manual update signal
			select {
			case <-timer.C:
			case <-sendWeatherCh:
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// getLocation returns the latitude and longitude
// given a location
func getLocation(ctx context.Context, loc string) (lat, lon float64, err error) {
	// Create request URL and perform GET request
	reqURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search.php?q=%s&format=jsonv2", url.QueryEscape(loc))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	// Decode JSON from response into OSMData
	data := OSMData{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return
	}
	// If no data points
	if len(data) == 0 {
		return
	}

	// Get first data point
	out := data[0]

	// Attempt to parse latitude
	lat, err = strconv.ParseFloat(out.Lat, 64)
	if err != nil {
		return
	}
	// Attempt to parse longitude
	lon, err = strconv.ParseFloat(out.Lon, 64)
	if err != nil {
		return
	}

	return
}

// getWeather gets weather data given a latitude and longitude
func getWeather(ctx context.Context, lat, lon float64) (*METResponse, error) {
	// Create new GET request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			"https://api.met.no/weatherapi/locationforecast/2.0/complete?lat=%.2f&lon=%.2f",
			lat,
			lon,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Set identifying user agent as per NMI requirements
	req.Header.Set("User-Agent", fmt.Sprintf("ITD/%s gitea.arsenm.dev/Arsen6331/itd", strings.TrimSpace(version)))

	// Perform request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Decode JSON from response to METResponse struct
	out := &METResponse{}
	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// parseSymbol determines what weather icon a symbol code codes for.
func parseSymbol(symCode string) infinitime.WeatherIcon {
	switch {
	case strings.Contains(symCode, "lightrain"):
		return infinitime.WeatherIconRain
	case strings.Contains(symCode, "rain"):
		return infinitime.WeatherIconCloudsWithRain
	case strings.Contains(symCode, "snow"),
		strings.Contains(symCode, "sleet"):
		return infinitime.WeatherIconSnow
	case strings.Contains(symCode, "thunder"):
		return infinitime.WeatherIconThunderstorm
	default:
		return infinitime.WeatherIconClear
	}
}
