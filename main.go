package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"teslamate-mqtt-api/config"
	"teslamate-mqtt-api/datapoints"
	"teslamate-mqtt-api/handlers"
	"teslamate-mqtt-api/msg"
	"teslamate-mqtt-api/store"
)

func main() {
	cfg := config.NewConfig()
	values := store.NewStore()
	topicPrefix := "teslamate/cars/"

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	httpLogger := httplog.NewLogger("http", httplog.Options{
		Concise: true,
	})

	datapoints.ValidDatapoints = datapoints.CalculateValidDatapoints(cfg.ApiKey != "")

	m := msg.NewMsg(cfg, topicPrefix, &values)
	err := m.Connect()
	if err != nil {
		log.Fatal().Err(err).Str("service", "mqtt").Msgf("Error connecting...", err)
	}

	h := handlers.NewCarsHandler(&values, cfg)
	auth := handlers.NewAuthMiddleware(cfg.ApiKey)

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(httpLogger))
	r.Use(auth.ApiKeyAuth())

	r.Get("/cars", h.Cars)
	r.Get("/car/{id}", h.Car)

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Error().Err(err).Str("service", "http")
	}

	m.Disconnect()
}
