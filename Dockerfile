# syntax = docker/dockerfile:1.4

FROM alpine:3.19.1@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b as RUN

# Add runtime dependencies
# - tzdata: Go time required external dependency eg: TRANSIP and possibly others
# - ca-certificates: Needed for https to work properly
RUN apk update && apk add --no-cache tzdata ca-certificates && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
