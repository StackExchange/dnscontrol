package svg // import "github.com/tdewolff/minify/svg"

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestPathData(t *testing.T) {
	var pathDataTests = []struct {
		pathData string
		expected string
	}{
		{"M10 10 20 10", "M10 10H20"},
		{"M10 10 10 20", "M10 10V20"},
		{"M50 50 100 100", "M50 50l50 50"},
		{"m50 50 40 40m50 50", "m50 50 40 40m50 50"},
		{"M10 10zM15 15", "M10 10zm5 5"},
		{"M50 50H55V55", "M50 50h5v5"},
		{"M10 10L11 10 11 11", "M10 10h1v1"},
		{"M10 10l1 0 0 1", "M10 10h1v1"},
		{"M10 10L11 11 0 0", "M10 10l1 1L0 0"},
		{"M246.614 51.028L246.614-5.665 189.922-5.665", "M246.614 51.028V-5.665H189.922"},
		{"M100,200 C100,100 250,100 250,200 S400,300 400,200", "M1e2 2e2c0-1e2 150-1e2 150 0s150 1e2 150 0"},
		{"M200,300 Q400,50 600,300 T1000,300", "M2e2 3e2q2e2-250 4e2.0t4e2.0"},
		{"M300,200 h-150 a150,150 0 1,0 150,-150 z", "M3e2 2e2H150A150 150 0 1 0 3e2 50z"},
		{"x5 5L10 10", "L10 10"},

		{"M.0.1", "M0 .1"},
		{"M200.0.1", "M2e2.1"},
		{"M0 0a3.28 3.28.0.0.0 3.279 3.28", "M0 0a3.28 3.28.0 0 0 3.279 3.28"}, // #114
		{"A1.1.0.0.0.0.2.3", "A1.1.0.0 0 0 .2."},                               // bad input (sweep and large-arc are not booleans) gives bad output

		// fuzz
		{"", ""},
		{"ML", ""},
		{".8.00c0", ""},
		{".1.04h0e6.0e6.0e0.0", "h0 0 0 0"},
		{"M.1.0.0.2Z", "M.1.0.0.2z"},
		{"A.0.0.0.0.3.2e3.7.0.0.0.0.0.1.3.0.0.0.0.2.3.2.0.0.0.0.20.2e-10.0.0.0.0.0.0.0.0", "A0 0 0 0 .3 2e2.7.0.0.0 0 0 .1.3 30 0 0 0 .2.3.2 3 20 0 0 .2 2e-1100 11 0 0 0 "}, // bad input (sweep and large-arc are not booleans) gives bad output
	}

	p := NewPathData(&Minifier{Decimals: -1})
	for _, tt := range pathDataTests {
		t.Run(tt.pathData, func(t *testing.T) {
			path := p.ShortenPathData([]byte(tt.pathData))
			test.Minify(t, tt.pathData, nil, string(path), tt.expected)
		})
	}
}

////////////////////////////////////////////////////////////////

func BenchmarkShortenPathData(b *testing.B) {
	p := NewPathData(&Minifier{})
	r := []byte("M8.64,223.948c0,0,143.468,3.431,185.777-181.808c2.673-11.702-1.23-20.154,1.316-33.146h16.287c0,0-3.14,17.248,1.095,30.848c21.392,68.692-4.179,242.343-204.227,196.59L8.64,223.948z")
	for i := 0; i < b.N; i++ {
		p.ShortenPathData(r)
	}
}
