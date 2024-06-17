package cloudflare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination_Done(t *testing.T) {
	testCases := map[string]struct {
		r        ResultInfo
		expected bool
	}{
		"missing ResultInfo pagination information": {
			r:        ResultInfo{Page: 1},
			expected: true,
		},
		"total pages greater than page": {
			r:        ResultInfo{Page: 1, TotalPages: 2},
			expected: false,
		},
		"total pages greater than page (alot)": {
			r:        ResultInfo{Page: 1, TotalPages: 200},
			expected: false,
		},
		// this should never happen
		"total pages less than page": {
			r:        ResultInfo{Page: 3, TotalPages: 1},
			expected: true,
		},
		"total pages missing but done": {
			r:        ResultInfo{Page: 4, Total: 70, PerPage: 25},
			expected: true,
		},
		"total pages missing and not done": {
			r:        ResultInfo{Page: 1, Total: 70, PerPage: 25},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.r.Done(), tc.expected)
		})
	}
}

func TestPagination_Next(t *testing.T) {
	testCases := map[string]struct {
		r        ResultInfo
		expected ResultInfo
	}{
		"missing ResultInfo pagination information": {
			r:        ResultInfo{Page: 1},
			expected: ResultInfo{Page: 1},
		},
		"total pages greater than page": {
			r:        ResultInfo{Page: 1, TotalPages: 3},
			expected: ResultInfo{Page: 2, TotalPages: 3},
		},
		"total pages greater than page (alot)": {
			r:        ResultInfo{Page: 1, TotalPages: 3000},
			expected: ResultInfo{Page: 2, TotalPages: 3000},
		},
		// bug, this should never happen
		"total pages less than page": {
			r:        ResultInfo{Page: 3, TotalPages: 1},
			expected: ResultInfo{Page: 3, TotalPages: 1},
		},
		"use per page and greater than page": {
			r:        ResultInfo{Page: 4, Total: 70, PerPage: 25},
			expected: ResultInfo{Page: 4, Total: 70, PerPage: 25},
		},
		"use per page and less than page": {
			r:        ResultInfo{Page: 1, Total: 70, PerPage: 25},
			expected: ResultInfo{Page: 2, Total: 70, PerPage: 25},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			assert.Equal(t, tc.r.Next(), tc.expected)
		})
	}
}

func TestPagination_HasMorePages(t *testing.T) {
	testCases := map[string]struct {
		r        ResultInfo
		expected bool
	}{
		"missing ResultInfo pagination information": {
			r:        ResultInfo{Page: 1},
			expected: false,
		},
		"total pages greater than page": {
			r:        ResultInfo{Page: 1, TotalPages: 3},
			expected: true,
		},
		"total pages greater than page (alot)": {
			r:        ResultInfo{Page: 1, TotalPages: 3000},
			expected: true,
		},
		// bug, this should never happen
		"total pages less than page": {
			r:        ResultInfo{Page: 3, TotalPages: 1},
			expected: false,
		},
		"use per page and greater than page": {
			r:        ResultInfo{Page: 4, Total: 70, PerPage: 25},
			expected: false,
		},
		"use per page and less than page": {
			r:        ResultInfo{Page: 1, Total: 70, PerPage: 25},
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			assert.Equal(t, tc.r.HasMorePages(), tc.expected)
		})
	}
}
