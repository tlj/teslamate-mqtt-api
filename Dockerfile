FROM golang:alpine as builder

RUN mkdir /build

ADD . /build/

WORKDIR /build

RUN go build -o tesla-mqtt-api .

# PRODUCTION
FROM alpine

RUN adduser -S -D -H -h /app appuser

USER appuser

COPY --from=builder /build/tesla-mqtt-api /app/

WORKDIR /app

EXPOSE 3000

ENTRYPOINT ["./tesla-mqtt-api"]
CMD [""]