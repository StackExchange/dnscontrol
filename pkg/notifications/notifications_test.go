package notifications

import "testing"

func Test_stripAnsiColorsValid(t *testing.T) {

	coloredStr := "\x1b[0133myellow\x1b[0m" // 33 == yellow
	nonColoredStr := "yellow"

	s := stripAnsiColors(coloredStr)
	if s != nonColoredStr {
		t.Errorf("stripAnsiColors() stripped %q different from %q", coloredStr, nonColoredStr)
	}
}

func Test_stripAnsiColorsInvalid(t *testing.T) {

	coloredStr := "\x1b[01AAmyellow\x1b[0m" // AA not a real color
	nonColoredStr := "yellow"

	s := stripAnsiColors(coloredStr)
	if s == nonColoredStr {
		t.Errorf("stripAnsiColors() stripped %q should be different from %q", coloredStr, nonColoredStr)
	}
}
