# syntax = docker/dockerfile:1.4

FROM alpine:3.17.0@sha256:8914eb54f968791faf6a8638949e480fef81e697984fba772b3976835194c6d4 as RUN

#RUN --mount=type=cache,target=/var/cache/apk \
#    apk update \
#    && apk add ca-certificates \
#    && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
