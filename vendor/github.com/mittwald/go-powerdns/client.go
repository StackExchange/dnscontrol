package pdns

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mittwald/go-powerdns/apis/cache"
	"github.com/mittwald/go-powerdns/apis/search"
	"github.com/mittwald/go-powerdns/apis/servers"
	"github.com/mittwald/go-powerdns/apis/zones"
	"github.com/mittwald/go-powerdns/pdnshttp"
)

type client struct {
	baseURL       string
	httpClient    *http.Client
	authenticator pdnshttp.ClientAuthenticator
	debugOutput   io.Writer

	cache   cache.Client
	search  search.Client
	servers servers.Client
	zones   zones.Client
}

type ClientOption func(c *client) error

// New creates a new PowerDNS client. Various client options can be used to configure
// the PowerDNS client (see examples).
func New(opt ...ClientOption) (Client, error) {
	c := client{
		baseURL:       "http://localhost:8081",
		httpClient:    http.DefaultClient,
		debugOutput:   ioutil.Discard,
		authenticator: &pdnshttp.NoopAuthenticator{},
	}

	for i := range opt {
		if err := opt[i](&c); err != nil {
			return nil, err
		}
	}

	if c.authenticator != nil {
		err := c.authenticator.OnConnect(c.httpClient)
		if err != nil {
			return nil, err
		}
	}

	hc := pdnshttp.NewClient(c.baseURL, c.httpClient, c.authenticator, c.debugOutput)

	c.servers = servers.New(hc)
	c.zones = zones.New(hc)
	c.search = search.New(hc)
	c.cache = cache.New(hc)

	return &c, nil
}

func (c *client) Status() error {
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return err
	}

	if err := c.authenticator.OnRequest(req); err != nil {
		return err
	}

	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) WaitUntilUp(ctx context.Context) error {
	up := make(chan error)
	cancel := false

	go func() {
		for !cancel {
			req, err := http.NewRequest("GET", c.baseURL, nil)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			_, err = c.httpClient.Do(req)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			up <- nil
			return
		}
	}()

	select {
	case <-up:
		return nil
	case <-ctx.Done():
		cancel = true
		return errors.New("context exceeded")
	}
}

func (c *client) Servers() servers.Client {
	return c.servers
}

func (c *client) Zones() zones.Client {
	return c.zones
}

func (c *client) Search() search.Client {
	return c.search
}

func (c *client) Cache() cache.Client {
	return c.cache
}
