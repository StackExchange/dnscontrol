FROM golang:1.14-alpine AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN apk update && apk add git
RUN GO111MODULE=on go run build/build.go -os=linux
RUN cp dnscontrol-Linux /go/bin/dnscontrol
RUN dnscontrol version
RUN go build -o cmd/convertzone/convertzone cmd/convertzone/main.go
RUN cp cmd/convertzone/convertzone /go/bin/convertzone

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build-env /go/bin/dnscontrol /usr/local/bin
COPY --from=build-env /go/bin/convertzone /usr/local/bin
WORKDIR /dns
RUN dnscontrol version
CMD dnscontrol
