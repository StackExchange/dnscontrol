package bind

import (
	"testing"
	"time"
)

func Test_generate_serial_1(t *testing.T) {
	d1, _ := time.Parse("20060102", "20150108")
	d4, _ := time.Parse("20060102", "40150108")
	d12, _ := time.Parse("20060102", "20151231")
	var tests = []struct {
		Given    uint32
		Today    time.Time
		Expected uint32
	}{
		{0, d1, 2015010800},
		{1, d1, 2015010800},
		{123, d1, 2015010800},
		{2015010800, d1, 2015010801},
		{2015010801, d1, 2015010802},
		{2015010802, d1, 2015010803},
		{2015010898, d1, 2015010899},
		{2015010899, d1, 2015010900},
		{2015090401, d1, 2015090402},
		{201509040, d1, 2015010800},
		{20150904, d1, 2015010800},
		{2015090, d1, 2015010800},
		// If the number is very large, just increment:
		{2099000000, d1, 2099000001},
		// Verify 32-bits is enough to carry us 200 years in the future:
		{4015090401, d4, 4015090402},
		// Verify Dec 31 edge-case:
		{2015123099, d12, 2015123100},
		{2015123100, d12, 2015123101},
		{2015123101, d12, 2015123102},
		{2015123102, d12, 2015123103},
		{2015123198, d12, 2015123199},
		{2015123199, d12, 2015123200},
		{2015123200, d12, 2015123201},
		{201512310, d12, 2015123100},
	}

	for i, tst := range tests {
		expected := tst.Expected
		nowFunc = func() time.Time {
			return tst.Today
		}
		found := generateSerial(tst.Given)
		if expected != found {
			t.Fatalf("Test:%d/%v: Expected (%d) got (%d)\n", i, tst.Given, expected, found)
		}
	}
}
