# syntax = docker/dockerfile:1.4

FROM alpine:3.17.1@sha256:f271e74b17ced29b915d351685fd4644785c6d1559dd1f2d4189a5e851ef753a as RUN

#RUN --mount=type=cache,target=/var/cache/apk \
#    apk update \
#    && apk add ca-certificates \
#    && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
