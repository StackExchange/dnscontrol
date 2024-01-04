package realtimeregister

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveEscapeChars(t *testing.T) {
	cleanedString := removeEscapeChars("\\\\\\\"")
	assert.Equal(t, "\\\"", cleanedString)
}

func TestAddEscapeChars(t *testing.T) {
	cleanedString := addEscapeChars("\\\"")
	assert.Equal(t, "\\\\\\\"", cleanedString)
}
