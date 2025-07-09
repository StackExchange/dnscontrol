// Copyright (c) 2017 Gorillalabs. All rights reserved.

package utils

import "testing"

func TestRandomStrings(t *testing.T) {
	r1 := CreateRandomString(8)
	r2 := CreateRandomString(8)

	if r1 == r2 {
		t.Error("Failed to create random strings: The two generated strings are identical.")
	} else if len(r1) != 16 {
		t.Errorf("Expected the random string to contain 16 characters, but got %d.", len(r1))
	}
}
