# syntax = docker/dockerfile:1.4

FROM alpine:3.20.2@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5 as RUN

# Add runtime dependencies
# - tzdata: Go time required external dependency eg: TRANSIP and possibly others
# - ca-certificates: Needed for https to work properly
RUN apk update && apk add --no-cache tzdata ca-certificates && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
