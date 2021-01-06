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
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	msgTopic := strings.TrimPrefix(msg.Topic(), topicPrefix)

	carTopic := strings.Split(msgTopic, "/")
	car := carTopic[0]
	name := carTopic[1]

	if _, ok := values[car]; !ok {
		values[car] = make(map[string]interface{})
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

	// for tesladata-widget
	values[car]["Date"] = time.Now().Format(time.RFC3339)
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
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	sub(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v\n", err)
}

var (
	topicPrefix = "teslamate/cars/"
	values      map[string]map[string]interface{}
)

func sub(client mqtt.Client) {
	fmt.Printf("Subscribing to %s...\n", topicPrefix)
	topic := fmt.Sprintf("%s#", topicPrefix)
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
}

func main() {
	mqttHost := os.Getenv("MQTT_HOST")
	if mqttHost == "" {
		log.Fatal("MQTT_HOST environment variable is required.")
	}
	brokerDsn := fmt.Sprintf("tcp://%s:1883", mqttHost)

	values = make(map[string]map[string]interface{})
	values["response"] = nil

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerDsn)
	opts.SetClientID("teslamate-scriptable-api")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	log.Printf("Connecting to %s...\n", brokerDsn)
	client := mqtt.NewClient(opts)
	if conn := client.Connect(); conn.Wait() && conn.Error() != nil {
		panic(conn.Error())
	}

	r := chi.NewRouter()
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
