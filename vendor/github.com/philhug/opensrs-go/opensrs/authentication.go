package opensrs

import (
	"crypto/md5"
	"encoding/hex"
)

const (
	httpHeaderUserName  = "X-UserName"
	httpHeaderSignature = "X-Signature"
)

// Provides credentials that can be used for authenticating with OpenSRS.
//
type Credentials interface {
	// Returns the HTTP headers that should be set
	// to authenticate the HTTP Request.
	Headers(xml []byte) map[string]string
}

// API key MD5 authentication
type apiKeyMD5Credentials struct {
	userName string
	apiKey   string
}

// NewApiKeyMD5Credentials construct Credentials using the OpenSRS MD5 Api Key method.
func NewApiKeyMD5Credentials(userName string, apiKey string) Credentials {
	return &apiKeyMD5Credentials{userName: userName, apiKey: apiKey}
}

func (c *apiKeyMD5Credentials) Headers(xml []byte) map[string]string {
	h := md5.New()
	h.Write(xml)
	h.Write([]byte(c.apiKey))
	m := hex.EncodeToString(h.Sum(nil))

	h = md5.New()
	h.Write([]byte(m))
	h.Write([]byte(c.apiKey))
	m = hex.EncodeToString(h.Sum(nil))

	return map[string]string{httpHeaderUserName: c.userName, httpHeaderSignature: m}
}
