# go-namecheap

A Go library for using [the Namecheap API](https://www.namecheap.com/support/api/intro.aspx).

**Build Status:** [![Build Status](https://travis-ci.org/billputer/go-namecheap.png?branch=master)](https://travis-ci.org/billputer/go-namecheap)

## Examples

```go
package main
import (
  "fmt"
  namecheap "github.com/billputer/go-namecheap"
)

func main() {
  apiUser := "billwiens"
  apiToken := "xxxxxxx"
  userName := "billwiens"

  client := namecheap.NewClient(apiUser, apiToken, userName)

  // Get a list of your domains
  domains, _ := client.DomainsGetList()
  for _, domain := range domains {
    fmt.Printf("Domain: %+v\n\n", domain.Name)
  }

}
```

For more complete documentation, load up godoc and find the package.

## Development

- Source hosted at [GitHub](https://github.com/billputer/go-namecheap)
- Report issues and feature requests to [GitHub Issues](https://github.com/billputer/go-namecheap/issues)

Pull requests welcome!

## Attribution

Most concepts and code borrowed from the excellent [go-dnsimple](http://github.com/rubyist/go-dnsimple).
