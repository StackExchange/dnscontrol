// Copyright (c) 2017 Gorillalabs. All rights reserved.

package utils

import "testing"

func TestQuotingArguments(t *testing.T) {
	testcases := [][]string{
		{"", "''"},
		{"test", "'test'"},
		{"two words", "'two words'"},
		{"quo\"ted", "'quo\"ted'"},
		{"quo'ted", "'quo\"ted'"},
		{"quo\\'ted", "'quo\\\"ted'"},
		{"quo\"t'ed", "'quo\"t\"ed'"},
		{"es\\caped", "'es\\caped'"},
		{"es`caped", "'es`caped'"},
		{"es\\`caped", "'es\\`caped'"},
	}

	for i, testcase := range testcases {
		quoted := QuoteArg(testcase[0])

		if quoted != testcase[1] {
			t.Errorf("test %02d failed: input '%s', expected %s, actual %s", i+1, testcase[0], testcase[1], quoted)
		}
	}
}
