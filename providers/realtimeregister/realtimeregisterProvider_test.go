package realtimeregister

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveEscapeChars(t *testing.T) {
	cleanedString := removeEscapeChars("\\\\\\\"")
	assert.Equal(t, "\\\"", cleanedString)
}

func TestAddEscapeChars(t *testing.T) {
	addedString := addEscapeChars("\\\"")
	assert.Equal(t, "\\\\\\\"", addedString)
}
