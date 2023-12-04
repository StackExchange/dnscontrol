package transip

import (
	"testing"
)

func TestRemoveSlashes(t *testing.T) {

	data := [][]string{
		{
			`quote"d`, `quote"d`,
			`quote\"d`, `quote"d`,
			`quote\\"d`, `quote\"d`,
			`m\o\\r\\\\e`, `mo\r\\e`,
		},
	}

	for _, testCase := range data {
		result := removeSlashes(testCase[0])
		if result != testCase[1] {
			t.Fatalf(`Failed on "%s". Expected "%s"; got "%s".`, testCase[0], testCase[1], result)
		}
	}

}
