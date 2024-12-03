FROM golang:1.23 AS build
RUN go build -v std

RUN apt-get update
RUN apt-get -y install libopus-dev libopusfile-dev

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal


RUN go build -o /konfa-voice ./cmd/server/main.go 

# run container
FROM debian:stable-slim

RUN apt-get update
RUN apt-get -y install libopus0 libopusfile0
#Adding root serts for ssl
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /konfa-voice /app/konfa-voice

WORKDIR /app

ENTRYPOINT [ "/app/konfa-voice" ]