package msg

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"math"
	"strconv"
	"strings"
	"teslamate-mqtt-api/config"
	"teslamate-mqtt-api/datapoints"
	"teslamate-mqtt-api/store"
	"time"
)

type Msg struct {
	cfg         *config.Config
	client      mqtt.Client
	topicPrefix string
	store       *store.Store
}

func (m *Msg) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.cfg.BrokerDsn)
	opts.SetClientID(m.cfg.ClientID)
	opts.SetDefaultPublishHandler(m.MessageHandler())
	opts.OnConnect = m.ConnectHandler()
	opts.OnConnectionLost = m.ConnectionLostHandler()

	log.Info().Msgf("Connecting to %s...", m.cfg.BrokerDsn)
	m.client = mqtt.NewClient(opts)

	if conn := m.client.Connect(); conn.Wait() && conn.Error() != nil {
		return conn.Error()
	}

	return nil
}

func (m *Msg) Disconnect() {
	m.client.Disconnect(250)
}

func NewMsg(cfg *config.Config, topicPrefix string, store *store.Store) *Msg {
	return &Msg{
		cfg:         cfg,
		topicPrefix: topicPrefix,
		store:       store,
	}
}

func (m *Msg) sub(client mqtt.Client) {
	log.Info().Msgf("Subscribing to %s...", m.topicPrefix)
	topic := fmt.Sprintf("%s#", m.topicPrefix)
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
}

func (m *Msg) ConnectionLostHandler() mqtt.ConnectionLostHandler {
	return func(client mqtt.Client, err error) {
		log.Error().Err(err).Str("service", "mqtt").Msg("Connect lost")
	}
}

func (m *Msg) ConnectHandler() mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		m.sub(client)
	}
}

func (m *Msg) MessageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		msgTopic := strings.TrimPrefix(msg.Topic(), m.topicPrefix)

		carTopic := strings.Split(msgTopic, "/")
		car := carTopic[0]
		name := carTopic[1]

		if _, ok := (*m.store)[car]; !ok {
			(*m.store)[car] = make(map[string]interface{})
			(*m.store)[car]["measure"] = m.cfg.DistanceUnit
			(*m.store)[car]["temperature"] = m.cfg.TemperatureUnit
		}

		if _, ok := datapoints.ValidDatapoints[name]; !ok {
			return
		}

		if string(msg.Payload()) == "false" {
			(*m.store)[car][name] = 0
		} else if string(msg.Payload()) == "true" {
			(*m.store)[car][name] = 1
		} else if intVal, err := strconv.Atoi(string(msg.Payload())); err == nil {
			(*m.store)[car][name] = intVal
		} else if floatVal, err := strconv.ParseFloat(string(msg.Payload()), 64); err == nil {
			(*m.store)[car][name] = floatVal
		} else {
			(*m.store)[car][name] = string(msg.Payload())
		}

		// for tesladata-widget we need to do some transformations

		(*m.store)[car]["Date"] = time.Now().Format(time.RFC3339)

		if strings.HasSuffix(name, "_km") {
			if m.cfg.DistanceUnit == config.DistanceUnitKm {
				(*m.store)[car]["measure"] = "km"
				(*m.store)[car][strings.TrimSuffix(name, "_km")] = (*m.store)[car][name]
			} else {
				(*m.store)[car]["measure"] = "imperial"
				(*m.store)[car][strings.TrimSuffix(name, "_km")] = math.Round(((*m.store)[car][name].(float64)/1.609)*100) / 100
			}
		}

		if name == "state" {
			switch string(msg.Payload()) {
			case "asleep":
				(*m.store)[car]["carState"] = "Sleeping"
			case "suspended":
				(*m.store)[car]["carState"] = "Sleeping"
			case "online":
				(*m.store)[car]["carState"] = "Idle"
			case "charging":
				(*m.store)[car]["carState"] = "Charging"
			case "driving":
				(*m.store)[car]["carState"] = "Driving"
			default:
				(*m.store)[car]["carState"] = string(msg.Payload())
			}
		}

		if name == "rated_battery_range_km" {
			(*m.store)[car]["battery_range"] = (*m.store)[car]["rated_battery_range_km"]
		}

		if name == "inside_temp" {
			(*m.store)[car]["inside_tempF"] = math.Round((((*m.store)[car]["inside_temp"].(float64)*9/5)+32)*100) / 100
		}
	}
}
