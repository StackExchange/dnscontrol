# syntax = docker/dockerfile:1.4

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 as RUN

# Add runtime dependencies
# - tzdata: Go time required external dependency eg: TRANSIP and possibly others
# - ca-certificates: Needed for https to work properly
RUN apk update && apk add --no-cache tzdata ca-certificates && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
