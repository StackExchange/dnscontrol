# cloudflare-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cloudflare/cloudflare-go.svg)](https://pkg.go.dev/github.com/cloudflare/cloudflare-go)
![Test](https://github.com/cloudflare/cloudflare-go/workflows/Test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudflare/cloudflare-go?style=flat-square)](https://goreportcard.com/report/github.com/cloudflare/cloudflare-go)

> **Note**: This library is under active development as we expand it to cover
> our (expanding!) API. Consider the public API of this package a little
> unstable as we work towards a v1.0.

A Go library for interacting with
[Cloudflare's API v4](https://api.cloudflare.com/). This library allows you to:

- Manage and automate changes to your DNS records within Cloudflare
- Manage and automate changes to your zones (domains) on Cloudflare, including
  adding new zones to your account
- List and modify the status of WAF (Web Application Firewall) rules for your
  zones
- Fetch Cloudflare's IP ranges for automating your firewall whitelisting

A command-line client, [flarectl](cmd/flarectl), is also available as part of
this project.

## Installation

You need a working Go environment. We officially support only currently supported Go versions according to [Go project's release policy](https://go.dev/doc/devel/release#policy).

```
go get github.com/cloudflare/cloudflare-go
```

## Getting Started

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	// Construct a new API object using a global API key
	api, err := cloudflare.New(os.Getenv("CLOUDFLARE_API_KEY"), os.Getenv("CLOUDFLARE_API_EMAIL"))
	// alternatively, you can use a scoped API token
	// api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	// Most API calls require a Context
	ctx := context.Background()

	// Fetch user details on the account
	u, err := api.UserDetails(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// Print user details
	fmt.Println(u)
}
```

Also refer to the
[API documentation](https://pkg.go.dev/github.com/cloudflare/cloudflare-go) for
how to use this package in-depth.

## Experimental improvements

This library is starting to ship with experimental improvements that are not yet
ready for production but will be introduced before the next major version. See
[experimental README](/docs/experimental.md) for full details.

## Contributing

Pull Requests are welcome, but please open an issue (or comment in an existing
issue) to discuss any non-trivial changes before submitting code.

## License

BSD licensed. See the [LICENSE](LICENSE) file for details.
