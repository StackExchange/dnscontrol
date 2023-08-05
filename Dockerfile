# syntax = docker/dockerfile:1.4

FROM alpine:3.18.2@sha256:02bb6f428431fbc2809c5d1b41eab5a68350194fb508869a33cb1af4444c9b11 as RUN

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
