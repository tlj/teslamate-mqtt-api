package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thanhpk/randstr"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	msgTopic := strings.TrimPrefix(msg.Topic(), topicPrefix)

	carTopic := strings.Split(msgTopic, "/")
	car := carTopic[0]
	name := carTopic[1]

	if _, ok := values[car]; !ok {
		values[car] = make(map[string]interface{})
	}

	if _, ok := validDatapoints[name]; !ok {
		return
	}

	if string(msg.Payload()) == "false" {
		values[car][name] = 0
	} else if string(msg.Payload()) == "true" {
		values[car][name] = 1
	} else if intVal, err := strconv.Atoi(string(msg.Payload())); err == nil {
		values[car][name] = intVal
	} else if floatVal, err := strconv.ParseFloat(string(msg.Payload()), 64); err == nil {
		values[car][name] = floatVal
	} else {
		values[car][name] = string(msg.Payload())
	}

	// for tesladata-widget we need to do some transformations

	values[car]["Date"] = time.Now().Format(time.RFC3339)

	if strings.HasSuffix(name, "_km") {
		values[car]["measure"] = "km"
		values[car][strings.TrimSuffix(name, "_km")] = values[car][name]
	}

	if name == "state" {
		switch string(msg.Payload()) {
		case "asleep":
			values[car]["carState"] = "Sleeping"
		case "suspended":
			values[car]["carState"] = "Sleeping"
		case "online":
			values[car]["carState"] = "Idle"
		case "charging":
			values[car]["carState"] = "Charging"
		case "driving":
			values[car]["carState"] = "Driving"
		default:
			values[car]["carState"] = string(msg.Payload())
		}
	}

	if name == "rated_battery_range_km" {
		values[car]["battery_range"] = values[car]["rated_battery_range_km"]
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	sub(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v\n", err)
}

var (
	topicPrefix     = "teslamate/cars/"
	values          map[string]map[string]interface{}
	validDatapoints map[string]bool

	basicDatapoints = map[string]bool{
		"battery_level":          true,
		"charge_energy_added":    true,
		"charge_limit_soc":       true,
		"display_name":           true,
		"est_battery_range_km":   true,
		"exterior_color":         true,
		"ideal_battery_range_km": true,
		"inside_temp":            true,
		"is_climate_on":          true,
		"is_preconditioning":     true,
		"outside_temp":           true,
		"model":                  true,
		"plugged_in":             true,
		"rated_battery_range_km": true,
		"spoiler_type":           true,
		"state":                  true,
		"time_to_full_charge":    true,
		"update_available":       true,
		"update_version":         true,
		"usable_battery_level":   true,
		"version":                true,
		"wheel_type":             true,
	}

	authenticatedDatapoints = map[string]bool{
		"doors_open":      true,
		"elevation":       true,
		"is_user_present": true,
		"latitude":        true,
		"longitude":       true,
		"locked":          true,
		"odometer":        true,
		"sentry_mode":     true,
		"speed":           true,
		"trunk_open":      true,
	}
)

func sub(client mqtt.Client) {
	log.Printf("Subscribing to %s...\n", topicPrefix)
	topic := fmt.Sprintf("%s#", topicPrefix)
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
}

func apiKeyAuth(requiredApiKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.URL.Query().Get("api_key")
			if apiKey != requiredApiKey {
				w.Write([]byte(`{"response":"invalid or missing api_key"}`))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	mqttHost := os.Getenv("MQTT_HOST")
	if mqttHost == "" {
		log.Fatal("MQTT_HOST environment variable is required.")
	}
	brokerDsn := fmt.Sprintf("tcp://%s:1883", mqttHost)

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Println("Running without API_KEY, limiting the exposed data.")
	}

	clientID := fmt.Sprintf("teslamate-mqtt-api-%s", randstr.Hex(6))
	log.Printf("Connecting with clientID %s", clientID)

	values = make(map[string]map[string]interface{})
	values["response"] = nil

	validDatapoints = make(map[string]bool)
	for k, v := range basicDatapoints {
		validDatapoints[k] = v
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerDsn)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	log.Printf("Connecting to %s...\n", brokerDsn)
	client := mqtt.NewClient(opts)
	if conn := client.Connect(); conn.Wait() && conn.Error() != nil {
		panic(conn.Error())
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	if apiKey != "" {
		r.Use(apiKeyAuth(apiKey))
		for k, v := range authenticatedDatapoints {
			validDatapoints[k] = v
		}
	}
	r.Get("/car/{id}", func(w http.ResponseWriter, r *http.Request) {
		carID := chi.URLParam(r, "id")

		j, _ := json.Marshal(values[carID])
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	})

	r.Get("/cars", func(w http.ResponseWriter, r *http.Request) {
		cars := make(map[string]string)
		for id, car := range values {
			if car == nil {
				continue
			}
			cars[id] = car["display_name"].(string)
		}

		j, _ := json.Marshal(cars)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	})

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Print(err)
	}

	client.Disconnect(250)
}
