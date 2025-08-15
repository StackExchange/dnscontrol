package netx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

// NewTransportWithSOCKS5 builds an *http.Transport that optionally dials via SOCKS5.
// If socksURL is empty, it uses proxy.FromEnvironmentUsing(base) which understands
// ALL_PROXY/NO_PROXY (including socks5:// URLs).
func NewTransportWithSOCKS5(socksURL string) (*http.Transport, error) {
	base := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}

	var d proxy.Dialer

	if socksURL != "" {
		u, err := url.Parse(socksURL)
		if err != nil {
			return nil, err
		}
		var auth *proxy.Auth
		if u.User != nil {
			pw, _ := u.User.Password()
			auth = &proxy.Auth{User: u.User.Username(), Password: pw}
		}
		d, err = proxy.SOCKS5("tcp", u.Host, auth, base)
		if err != nil {
			return nil, err
		}
	} else {
		// Note: FromEnvironmentUsing returns only a Dialer (no error).
		d = proxy.FromEnvironmentUsing(base)
	}

	tr := &http.Transport{
		Proxy:               http.ProxyFromEnvironment, // HTTP proxies still honored
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:   true,
	}

	if cd, ok := d.(proxy.ContextDialer); ok {
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return cd.DialContext(ctx, network, addr)
		}
	} else {
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.Dial(network, addr)
		}
	}

	_ = http2.ConfigureTransport(tr)
	return tr, nil
}

// NewHTTPClientWithSOCKS5 returns an *http.Client with the transport above.
func NewHTTPClientWithSOCKS5(socksURL string) (*http.Client, error) {
	tr, err := NewTransportWithSOCKS5(socksURL)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: tr}, nil
}

// SOCKS5FromEnv returns DNSCONTROL_SOCKS5 if set, else "".
// If empty string is returned, NewTransportWithSOCKS5 will fall back to
// ALL_PROXY/NO_PROXY via proxy.FromEnvironmentUsing().
func SOCKS5FromEnv() string {
	if v := os.Getenv("DNSCONTROL_SOCKS5"); v != "" {
		return v
	}
	return ""
}
