FROM golang:1.24-bookworm AS build
RUN go build -v std

RUN apt-get update
RUN apt-get -y install libopus-dev libopusfile-dev

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN --mount=type=cache,mode=0777,target=/go/pkg/mod go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal


RUN --mount=type=cache,mode=0777,target=/go/pkg/mod \
    go build -o /confa-voice-relay ./cmd/server/main.go 

# run container
FROM debian:bookworm-slim

RUN apt-get update
RUN apt-get -y install libopus0 libopusfile0
#Adding root serts for ssl
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /confa-voice-relay /app/confa-voice-relay

WORKDIR /app

ENTRYPOINT [ "/app/confa-voice-relay" ]