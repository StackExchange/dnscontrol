package cloudflare

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceGroup(t *testing.T) {
	setup()
	defer teardown()

	key := "com.cloudflare.test.1"
	rg := NewResourceGroup(key)

	assert.Equal(t, rg.Name, key)
	assert.Equal(t, rg.Scope.Key, key)
}

func TestNewResourceGroupForAccount(t *testing.T) {
	setup()
	defer teardown()

	id := "some-fake-account-id"
	rg := NewResourceGroupForAccount(Account{
		ID: id,
	})

	key := fmt.Sprintf("com.cloudflare.api.account.%s", id)
	assert.Equal(t, rg.Name, key)
	assert.Equal(t, rg.Scope.Key, key)
}

func TestNewResourceGroupForZone(t *testing.T) {
	setup()
	defer teardown()

	id := "some-fake-zone-id"
	rg := NewResourceGroupForZone(Zone{
		ID: id,
	})

	key := fmt.Sprintf("com.cloudflare.api.account.zone.%s", id)
	assert.Equal(t, rg.Name, key)
	assert.Equal(t, rg.Scope.Key, key)
}
