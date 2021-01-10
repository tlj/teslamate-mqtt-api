
docker:
	docker buildx build --push --platform linux/arm/v7,linux/arm64/v8,linux/amd64 -t thomaslandro/teslamate-mqtt-api:latest .