# PowerDNS client library for Go

[![GoDoc](https://godoc.org/github.com/mittwald/go-powerdns?status.svg)](https://godoc.org/github.com/mittwald/go-powerdns)
[![Build Status](https://travis-ci.org/mittwald/go-powerdns.svg?branch=master)](https://travis-ci.org/mittwald/go-powerdns)
[![Maintainability](https://api.codeclimate.com/v1/badges/aa54a869f5ff56477a2a/maintainability)](https://codeclimate.com/github/mittwald/go-powerdns/maintainability)

This package contains a Go library for accessing the [PowerDNS][powerdns] Authoritative API.

## Supported features

- [x] Servers
- [x] Zones
- [ ] Cryptokeys
- [ ] Metadata
- [ ] TSIG Keys
- [x] Searching
- [ ] Statistics
- [x] Cache

## Installation

Install using `go get`:

```console
> go get github.com/mittwald/go-powerdns
```

## Usage

First, instantiate a client using `pdns.New`:

```go
client, err := pdns.New(
    pdns.WithBaseURL("http://localhost:8081"),
    pdns.WithAPIKeyAuthentication("supersecret"),
)
```

The client then offers more specialiced sub-clients, for example for managing server and zones.
Have a look at this library's [documentation][godoc] for more information.

## Complete example

```go
package main

import "context"
import "github.com/mittwald/go-powerdns"
import "github.com/mittwald/go-powerdns/apis/zones"

func main() {
    client, err := pdns.New(
        pdns.WithBaseURL("http://localhost:8081"),
        pdns.WithAPIKeyAuthentication("supersecret"),
    )
	
    if err != nil {
    	panic(err)
    }
    
    client.Zones().CreateZone(context.Background(), "localhost", zones.Zone{
        Name: "mydomain.example.",
        Type: zones.ZoneTypeZone,
        Kind: zones.ZoneKindNative,
        Nameservers: []string{
            "ns1.example.com.",
            "ns2.example.com.",
        },
        ResourceRecordSets: []zones.ResourceRecordSet{
            {Name: "foo.mydomain.example.", Type: "A", TTL: 60, Records: []zones.Record{{Content: "127.0.0.1"}}},
        },
    })
}
```

[powerdns]: https://github.com/PowerDNS/pdns
[godoc]: https://godoc.org/github.com/mittwald/go-powerdns
