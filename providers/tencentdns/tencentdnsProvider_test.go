package tencentdns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTencentDNS(t *testing.T) {
	config := map[string]string{
		"secret_id":  "test-id",
		"secret_key": "test-key",
		"region":     "ap-guangzhou",
	}

	provider, err := newTencentDNS(config)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.client)
}

func TestNewTencentDNS_MissingCreds(t *testing.T) {
	config := map[string]string{
		"secret_id": "test-id",
		// missing secret_key
	}

	provider, err := newTencentDNS(config)
	assert.Error(t, err)
	assert.Nil(t, provider)
}
