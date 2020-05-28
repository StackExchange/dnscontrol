package powerdns_test

import (
	"github.com/StackExchange/dnscontrol/v3/providers/powerdns"
	"testing"
)

func TestEnsureDotSuffix(t *testing.T) {
	expected := "example.org."
	actual := powerdns.EnsureDotSuffix("example.org")
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestConcatMessage(t *testing.T) {
	expected := "Message 1\n    Message 2"
	actual := powerdns.ConcatMessage([]string{"Message 1", "Message 2"})
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
