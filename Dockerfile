FROM golang:1.9 AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN go install .
RUN dnscontrol version

FROM ubuntu:xenial
COPY --from=build-env /go/bin/dnscontrol /usr/local/bin
WORKDIR /dns
RUN apt-get install -y ca-certificates
RUN dnscontrol version
CMD dnscontrol