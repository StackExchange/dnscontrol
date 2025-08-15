package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/netx"
)

func redactProxyURL(s string) string {
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return s
	}
	user := ""
	if u.User != nil && u.User.Username() != "" {
		user = u.User.Username() + "@"
	}
	// Drop password; print only scheme://[user@]host:port
	return u.Scheme + "://" + user + u.Host
}

// init runs before main() and installs a default HTTP client/transport that
// respects DNSCONTROL_SOCKS5 (ENV or defaults.js) or ALL_PROXY/NO_PROXY.
// Prints a one-line notice if a SOCKS5 proxy is configured.
func init() {
	socks := netx.SOCKS5FromEnv()
	src := ""
	if socks != "" {
		src = "DNSCONTROL_SOCKS5"
	} else {
		socks = netx.SOCKS5FromDefaultsJS()
		if socks != "" {
			src = "defaults.js"
		} else {
			// For banner only: if ALL_PROXY is socks5://..., show it.
			ap := os.Getenv("ALL_PROXY")
			if ap == "" {
				ap = os.Getenv("all_proxy")
			}
			if strings.HasPrefix(strings.ToLower(ap), "socks5://") {
				src = "ALL_PROXY"
				// Use ALL_PROXY value for both banner and transport.
				socks = ap
			}
		}
	}
	if tr, err := netx.NewTransportWithSOCKS5(socks); err == nil {
		http.DefaultTransport = tr
		http.DefaultClient = &http.Client{Transport: tr}
		if socks != "" {
			// Print to stderr to avoid polluting command output
			fmt.Fprintln(os.Stderr, "USING PROXY ("+src+"): "+redactProxyURL(socks))
		}
	}
}
