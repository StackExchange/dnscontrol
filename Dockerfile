# syntax = docker/dockerfile:1.4

FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715 as RUN

# Add runtime dependencies
# - tzdata: Go time required external dependency eg: TRANSIP and possibly others
# - ca-certificates: Needed for https to work properly
RUN apk update && apk add --no-cache tzdata ca-certificates && update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
