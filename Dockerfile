FROM golang:1.19.0-alpine3.16@sha256:0eb08c89ab1b0c638a9fe2780f7ae3ab18f6ecda2c76b908e09eb8073912045d AS build

WORKDIR /go/src/github.com/StackExchange/dnscontrol

ARG BUILD_VERSION

ENV GO111MODULE on

COPY . .

# build dnscontrol
RUN apk update \
    && apk add --no-cache ca-certificates curl gcc build-base git \
    && update-ca-certificates \
    && go build -v -trimpath -buildmode=pie -ldflags="-s -w -X main.SHA=${BUILD_VERSION}"

# Validation check
RUN cp dnscontrol /go/bin/dnscontrol
RUN dnscontrol version

# -----

FROM alpine:3.16.2@sha256:6c1b23807ddcee59b594ebe3cbbc36371fe9737f7a796881b3f72aeef81e6dcb

COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /go/bin/dnscontrol /usr/local/bin

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
