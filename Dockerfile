# syntax = docker/dockerfile:1.4

FROM alpine:3.18.2@sha256:25fad2a32ad1f6f510e528448ae1ec69a28ef81916a004d3629874104f8a7f70 as RUN

# Add runtime dependencies
# - tzdata: Go time required external dependency eg: TRANSIP and possibly others
# - ca-certificates: Needed for https to work properly
#RUN --mount=type=cache,target=/var/cache/apk \
#    apk update \
#    && apk add tzdata ca-certificates \
#    && update-ca-certificates
RUN echo TEST1
RUN --mount=type=cache,target=/var/cache/apk echo TEST2
RUN --mount=type=cache,target=/var/cache/apk apk update 
RUN --mount=type=cache,target=/var/cache/apk apk add --no-cache tzdata ca-certificates
RUN --mount=type=cache,target=/var/cache/apk update-ca-certificates

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
