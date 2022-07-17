package domainnameshop

import (
	"testing"
)

func TestFixTTL(t *testing.T) {
	for i, test := range []struct {
		given, expected uint32
	}{
		{1, 60},
		{TTLSteps*5 - 1, TTLSteps * 4},
		{maxAllowedTTL + 1, maxAllowedTTL},
	} {
		found := fixTTL(test.given)
		if found != test.expected {
			t.Errorf("Test %d: Expected %d, but was %d", i, test.expected, found)
		}
	}
}
