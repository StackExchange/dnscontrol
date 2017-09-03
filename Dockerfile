FROM golang:1.9 AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN go install .
RUN dnscontrol version

FROM alpine
WORKDIR /
COPY --from=build-env /go/bin/dnscontrol /dnscontrol
RUN ls -la
RUN pwd
RUN /dnscontrol version
ENTRYPOINT ./dnscontrol