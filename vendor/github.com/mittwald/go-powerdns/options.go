package pdns

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"io"
	"io/ioutil"
	"net/http"
)

// WithBaseURL sets a client's base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *client) error {
		c.baseURL = baseURL
		return nil
	}
}

// WithHTTPClient can be used to override a client's HTTP client.
// Otherwise, the default HTTP client will be used
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *client) error {
		c.httpClient = httpClient
		return nil
	}
}

// WithAPIKeyAuthentication adds API-key based authentication to the PowerDNS client.
// In effect, each HTTP request will have an additional header that contains the API key
// supplied to this function:
//  X-API-Key: {{ key }}
func WithAPIKeyAuthentication(key string) ClientOption {
	return func(c *client) error {
		c.authenticator = &pdnshttp.APIKeyAuthenticator{
			APIKey: key,
		}

		return nil
	}
}

// WithTLSAuthentication configures TLS-based authentication for the PowerDNS client.
// This is not a feature that is provided by PowerDNS natively, but might be implemented
// when the PowerDNS API is run behind a reverse proxy.
func WithTLSAuthentication(caFile string, clientCertFile string, clientKeyFile string) ClientOption {
	return func(c *client) error {
		cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return err
		}

		caBytes, err := ioutil.ReadFile(caFile)
		if err != nil {
			return err
		}

		ca, err := x509.ParseCertificates(caBytes)

		auth := pdnshttp.TLSClientCertificateAuthenticator{
			ClientCert: cert,
			ClientKey:  cert.PrivateKey,
			CACerts:    ca,
		}

		c.authenticator = &auth
		return nil
	}
}

// WithDebuggingOutput can be used to supply an io.Writer to the client into which all
// outgoing HTTP requests and their responses will be logged. Useful for debugging.
func WithDebuggingOutput(out io.Writer) ClientOption {
	return func(c *client) error {
		c.debugOutput = out
		return nil
	}
}
