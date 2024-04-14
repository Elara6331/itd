package infinitime

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

const (
	weatherVersion = 0

	currentWeatherType  = 0
	forecastWeatherType = 1
)

type WeatherIcon uint8

const (
	WeatherIconClear WeatherIcon = iota
	WeatherIconFewClouds
	WeatherIconClouds
	WeatherIconHeavyClouds
	WeatherIconCloudsWithRain
	WeatherIconRain
	WeatherIconThunderstorm
	WeatherIconSnow
	WeatherIconMist
)

// CurrentWeather represents the current weather
type CurrentWeather struct {
	Time        time.Time
	CurrentTemp float32
	MinTemp     float32
	MaxTemp     float32
	Location    string
	Icon        WeatherIcon
}

// Bytes returns the [CurrentWeather] struct encoded using the InfiniTime
// weather wire protocol.
func (cw CurrentWeather) Bytes() []byte {
	buf := &bytes.Buffer{}

	buf.WriteByte(currentWeatherType)
	buf.WriteByte(weatherVersion)

	_, offset := cw.Time.Zone()
	binary.Write(buf, binary.LittleEndian, cw.Time.Unix()+int64(offset))

	binary.Write(buf, binary.LittleEndian, int16(cw.CurrentTemp*100))
	binary.Write(buf, binary.LittleEndian, int16(cw.MinTemp*100))
	binary.Write(buf, binary.LittleEndian, int16(cw.MaxTemp*100))

	location := make([]byte, 32)
	copy(location, cw.Location)
	buf.Write(location)

	buf.WriteByte(byte(cw.Icon))

	return buf.Bytes()
}

// Forecast represents a weather forecast
type Forecast struct {
	Time time.Time
	Days []ForecastDay
}

// ForecastDay represents a forecast for a single day
type ForecastDay struct {
	MinTemp int16
	MaxTemp int16
	Icon    WeatherIcon
}

// Bytes returns the [Forecast] struct encoded using the InfiniTime
// weather wire protocol.
func (f Forecast) Bytes() []byte {
	buf := &bytes.Buffer{}

	buf.WriteByte(forecastWeatherType)
	buf.WriteByte(weatherVersion)

	_, offset := f.Time.Zone()
	binary.Write(buf, binary.LittleEndian, f.Time.Unix()+int64(offset))

	buf.WriteByte(uint8(len(f.Days)))

	for _, day := range f.Days {
		binary.Write(buf, binary.LittleEndian, day.MinTemp*100)
		binary.Write(buf, binary.LittleEndian, day.MaxTemp*100)
		buf.WriteByte(byte(day.Icon))
	}

	return buf.Bytes()
}

// SetCurrentWeather updates the current weather data on the PineTime
func (d *Device) SetCurrentWeather(cw CurrentWeather) error {
	c, err := d.getChar(weatherDataChar)
	if err != nil {
		return err
	}

	_, err = c.WriteWithoutResponse(cw.Bytes())
	return err
}

// SetForecast sets future forecast data on the PineTime
func (d *Device) SetForecast(f Forecast) error {
	c, err := d.getChar(weatherDataChar)
	if err != nil {
		return err
	}

	if len(f.Days) > 5 {
		return errors.New("amount of forecast days exceeds maximum of 5")
	}

	_, err = c.WriteWithoutResponse(f.Bytes())
	return err
}
