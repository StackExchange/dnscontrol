FROM golang:1.9 AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN go install .
RUN dnscontrol version

FROM alpine
COPY --from=build-env /go/bin/dnscontrol /usr/local/bin/
RUN ls -la /usr/local/bin/
RUN /usr/local/bin/dnscontrol version
ENTRYPOINT ./dnscontrol