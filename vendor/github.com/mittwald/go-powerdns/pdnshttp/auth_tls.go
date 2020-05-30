package pdnshttp

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
)

type TLSClientCertificateAuthenticator struct {
	CACerts    []*x509.Certificate
	ClientCert tls.Certificate
	ClientKey  crypto.PrivateKey
}

func (a *TLSClientCertificateAuthenticator) OnRequest(r *http.Request) error {
	return nil
}

func (a *TLSClientCertificateAuthenticator) OnConnect(c *http.Client) error {
	if c.Transport == nil {
		c.Transport = http.DefaultTransport
	}

	t, ok := c.Transport.(*http.Transport)
	if !ok {
		return fmt.Errorf("client.Transport is no *http.Transport, instead %t", c.Transport)
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}

	if t.TLSClientConfig.Certificates == nil {
		t.TLSClientConfig.Certificates = make([]tls.Certificate, 0, 1)
	}

	t.TLSClientConfig.Certificates = append(t.TLSClientConfig.Certificates, a.ClientCert)

	if t.TLSClientConfig.RootCAs == nil {
		systemPool, err := x509.SystemCertPool()
		if err != nil {
			return err
		}

		t.TLSClientConfig.RootCAs = systemPool
	}

	for i := range a.CACerts {
		t.TLSClientConfig.RootCAs.AddCert(a.CACerts[i])
	}

	return nil
}
