# syntax = docker/dockerfile:1.4

FROM alpine:3.17.3@sha256:124c7d2707904eea7431fffe91522a01e5a861a624ee31d03372cc1d138a3126 as RUN

#RUN --mount=type=cache,target=/var/cache/apk \
#    apk update \
#    && apk add ca-certificates \
#    && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
