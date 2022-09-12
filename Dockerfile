# syntax = docker/dockerfile:1.4

FROM alpine:3.16.2@sha256:bc41182d7ef5ffc53a40b044e725193bc10142a1243f395ee852a8d9730fc2ad as RUN

#RUN --mount=type=cache,target=/var/cache/apk \
#    apk update \
#    && apk add ca-certificates \
#    && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
