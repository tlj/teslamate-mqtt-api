# Teslamate MQTT API

Creates an API with all datapoints exposed from Teslamate to MQTT.

Endpoints:
* /cars
* /car/{id}

## Usage

Environment variables:

|name|description|
|----|-----------|
|MQTT_HOST| Usually the name of the mosquitto block in your docker-composer.yml |
|API_KEY| Setting your own api key here will expose more data through the api. Without the api key you will still get enough data to feed some things, like the TeslaData-Widget. |
|TEMPERATURE_UNIT| The unit of temperature ("C" or "F") |
|DISTANCE_UNIT| The unit for measuring distance ("km" or "imperial")|

Add to docker-compose.yml: 

    teslamate_mqtt_api:
      image: thomaslandro/teslamate-mqtt-api:latest
      container_name: teslamate_mqtt_api
      restart: always
      environment:
        - MQTT_HOST=mosquitto
        - API_KEY=define_your_own_here
        - TEMPERATURE_UNIT=C
        - DISTANCE_UNIT=km
      ports:
        - 3040:3000

## Use with TeslaData-Widget

Check the /cars endpoint to make a note of the id of the car you want to expose. 
Follow the instructions in TeslaData-Widget to use a custom API, and use this as your 
APIurl:

http(s)://{YOUR_HOST}:3040/car/{CAR_ID}?api_key={THE_ONE_DEFINED_IN_DOCKER_COMPOSER}

If you didn't set the API_KEY value in docker-composer.yml you can remove that part of the URL.
