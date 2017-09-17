FROM golang:1.9 AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN go run build/build.go
RUN cp dnscontrol-Linux /go/bin/dnscontrol
RUN dnscontrol version

FROM ubuntu:xenial
COPY --from=build-env /go/bin/dnscontrol /usr/local/bin
WORKDIR /dns
RUN apt-get update
RUN apt-get install -y ca-certificates
CMD dnscontrol