package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/weather"
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
			AirTemperature    float32 `json:"air_temperature"`
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
}

// OSMData represents lat/long data from
// OpenStreetMap Nominatim
type OSMData []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

var sendWeatherCh = make(chan struct{}, 1)

func initWeather(ctx context.Context, dev *infinitime.Device) error {
	if !k.Bool("weather.enabled") {
		return nil
	}

	// Get location based on string in config
	lat, lon, err := getLocation(ctx, k.String("weather.location"))
	if err != nil {
		return err
	}

	timer := time.NewTimer(time.Hour)

	go func() {
		for {
			// Attempt to get weather
			data, err := getWeather(ctx, lat, lon)
			if err != nil {
				log.Warn().Err(err).Msg("Error getting weather data")
				// Wait 15 minutes before retrying
				time.Sleep(15 * time.Minute)
				continue
			}

			// Get current data
			current := data.Properties.Timeseries[0]
			currentData := current.Data.Instant.Details

			// Add temperature event
			err = dev.AddWeatherEvent(weather.TemperatureEvent{
				TimelineHeader: weather.NewHeader(
					weather.EventTypeTemperature,
					time.Hour,
				),
				Temperature: int16(round(currentData.AirTemperature * 100)),
				DewPoint:    int16(round(currentData.DewPoint)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error adding temperature event")
			}

			// Add precipitation event
			err = dev.AddWeatherEvent(weather.PrecipitationEvent{
				TimelineHeader: weather.NewHeader(
					weather.EventTypePrecipitation,
					time.Hour,
				),
				Type:   parseSymbol(current.Data.NextHour.Summary.SymbolCode),
				Amount: uint8(round(current.Data.NextHour.Details.PrecipitationAmount)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error adding precipitation event")
			}

			// Add wind event
			err = dev.AddWeatherEvent(weather.WindEvent{
				TimelineHeader: weather.NewHeader(
					weather.EventTypeWind,
					time.Hour,
				),
				SpeedMin:     uint8(round(currentData.WindSpeed)),
				SpeedMax:     uint8(round(currentData.WindSpeed)),
				DirectionMin: uint8(round(currentData.WindDirection)),
				DirectionMax: uint8(round(currentData.WindDirection)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error adding wind event")
			}

			// Add cloud event
			err = dev.AddWeatherEvent(weather.CloudsEvent{
				TimelineHeader: weather.NewHeader(
					weather.EventTypeClouds,
					time.Hour,
				),
				Amount: uint8(round(currentData.CloudAreaFraction)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error adding clouds event")
			}

			// Add humidity event
			err = dev.AddWeatherEvent(weather.HumidityEvent{
				TimelineHeader: weather.NewHeader(
					weather.EventTypeHumidity,
					time.Hour,
				),
				Humidity: uint8(round(currentData.RelativeHumidity)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error adding humidity event")
			}

			// Add pressure event
			err = dev.AddWeatherEvent(weather.PressureEvent{
				TimelineHeader: weather.NewHeader(
					weather.EventTypePressure,
					time.Hour,
				),
				Pressure: int16(round(currentData.AirPressure)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error adding pressure event")
			}

			// Reset timer to 1 hour
			timer.Stop()
			timer.Reset(time.Hour)

			// Wait for timer to fire or manual update signal
			select {
			case <-timer.C:
			case <-sendWeatherCh:
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
	req.Header.Set("User-Agent", fmt.Sprintf("ITD/%s gitea.arsenm.dev/Arsen6331/itd", version))

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

// parseSymbol determines what type of precipitation a symbol code
// codes for.
func parseSymbol(symCode string) weather.PrecipitationType {
	switch {
	case strings.Contains(symCode, "lightrain"):
		return weather.PrecipitationTypeRain
	case strings.Contains(symCode, "rain"):
		return weather.PrecipitationTypeRain
	case strings.Contains(symCode, "snow"):
		return weather.PrecipitationTypeSnow
	case strings.Contains(symCode, "sleet"):
		return weather.PrecipitationTypeSleet
	case strings.Contains(symCode, "snow"):
		return weather.PrecipitationTypeSnow
	default:
		return weather.PrecipitationTypeNone
	}
}

// round rounds 32-bit floats to 32-bit integers
func round(f float32) int32 {
	return int32(math.Round(float64(f)))
}
