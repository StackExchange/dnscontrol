FROM golang:1.9 AS build-env
WORKDIR /go/src/github.com/StackExchange/dnscontrol
ADD . .
RUN go install .
RUN which dnscontrol
RUN dnscontrol version

FROM alpine
COPY --from=build-env /go/bin/dnscontrol /usr/local/bin/
RUN ls -la /usr/local/bin/
CMD /usr/local/bin/dnscontrol