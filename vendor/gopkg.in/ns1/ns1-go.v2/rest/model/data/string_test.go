package data

import (
	"fmt"
	"testing"
)

func BenchmarkToCamel(b *testing.B) {

	s := "this_is_an_awesome_string_to_test_camel_case. it_should_outperform_the_concatenated_version_substantially (it did)"

	for n := 0; n < b.N; n++ {
		ToCamel(s)
	}
}

func TestToCamel(t *testing.T) {
	s := "this_is_an_awesome_string_to_test_camel_case. it_should_cover_all_the_cases. even-ones-with-dashes."
	s = ToCamel(s)
	expected := "ThisIsAnAwesomeStringToTestCamelCaseItShouldCoverAllTheCasesEvenOnesWithDashes"
	if s != expected {
		t.Fatal(fmt.Sprintf("s should equal '%s', but was '%s'", expected, s))
	}
}
