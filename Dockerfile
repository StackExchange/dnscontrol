FROM golang:1.9-alpine AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN apk update && apk add git
RUN go run build/build.go -os=linux
RUN cp dnscontrol-Linux /go/bin/dnscontrol
RUN dnscontrol version

FROM alpine
COPY --from=build-env /go/bin/dnscontrol /usr/local/bin
WORKDIR /dns
RUN dnscontrol version
CMD dnscontrol