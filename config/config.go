package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/thanhpk/randstr"
	"os"
)

type Config struct {
	BrokerDsn       string
	ApiKey          string
	DistanceUnit    string
	TemperatureUnit string
	ClientID        string
}

type DistanceUnitType string
type TemperatureUnitType string

const (
	DistanceUnitKm            string = "km"
	DistanceUnitMiles         string = "imperial"
	TemperatureUnitCelsius    string = "C"
	TemperatureUnitFahrenheit string = "F"
)

func NewConfig() *Config {
	c := Config{}

	c.ClientID = fmt.Sprintf("teslamate-mqtt-api-%s", randstr.Hex(6))

	mqttHost := os.Getenv("MQTT_HOST")
	if mqttHost == "" {
		log.Fatal().Msg("MQTT_HOST environment variable is required.")
	}
	c.BrokerDsn = fmt.Sprintf("tcp://%s:1883", mqttHost)

	c.ApiKey = os.Getenv("API_KEY")
	c.DistanceUnit = os.Getenv("DISTANCE_UNIT")
	if c.DistanceUnit == "" || (c.DistanceUnit != DistanceUnitKm && c.DistanceUnit != DistanceUnitMiles) {
		c.DistanceUnit = DistanceUnitKm
		log.Info().Msgf("No DISTANCE_UNIT defined, using '%s' (alternative is '%s')...\n", c.DistanceUnit, DistanceUnitMiles)
	}

	c.TemperatureUnit = os.Getenv("TEMPERATURE_UNIT")
	if c.TemperatureUnit == "" || (c.TemperatureUnit != TemperatureUnitCelsius && c.TemperatureUnit != TemperatureUnitFahrenheit) {
		c.TemperatureUnit = TemperatureUnitCelsius
		log.Info().Msgf("No TEMPERATURE_UNIT defined, using '%s' (alternative is '%s')...\n", c.TemperatureUnit, TemperatureUnitFahrenheit)
	}

	return &c
}
