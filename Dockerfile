# syntax = docker/dockerfile:1.4

# We're using Debian because that is also the base image for the `golang` images
FROM debian:11@sha256:432f545c6ba13b79e2681f4cc4858788b0ab099fc1cca799cc0fae4687c69070 AS RUN

# Install `ca-certificates` since `debian` does not ship with them
# and they are required for a few of our supported providers
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY dnscontrol /usr/local/bin/

WORKDIR /dns

ENTRYPOINT ["/usr/local/bin/dnscontrol"]
