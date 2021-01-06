# Teslamate MQTT API

Creates an API with all datapoints exposed from Teslamate to MQTT.

Endpoints:
* /cars
* /car/{id}

## Usage

Add to docker-compose.yml: 

    teslamate_mqtt_api:
      image: thomaslandro/teslamate-mqtt-api:latest
      container_name: teslamate_mqtt_api
      restart: always
      environment:
        - MQTT_HOST=mosquitto
      ports:
        - 3040:3000

## Use with TeslaData-Widget

Check the /cars endpoint to make a note of the id of the car you want to expose. 
Follow the instructions in TeslaData-Widget to use a custom API, and use this as your 
APIurl:

http(s)://YOUR_HOST:3040/car/CAR_ID

