/*
	Copyright 2014 Google Inc. All rights reserved.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

// Package natsort provides an implementation of the natural sorting algorithm.
// See http://blog.codinghorror.com/sorting-for-humans-natural-sort-order/.
package natsort

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Strings sorts a slice of strings with Less.
func Strings(s []string) {
	sort.Sort(stringSlice(s))
}

// Less reports whether s is less than t.
func Less(s, t string) bool {
	return LessRunes([]rune(s), []rune(t))
}

// LessRunes reports whether s is less than t.
func LessRunes(s, t []rune) bool {
	nprefix := commonPrefix(s, t)
	if len(s) == nprefix && len(t) == nprefix {
		// equal
		return false
	}
	if len(s) == 0 || len(t) == 0 {
		return len(t) != 0
	}
	if strings.HasPrefix(string(s), "*") || strings.HasPrefix(string(t), "*") {
		if isDigit(t[0]) {
			return false
		}
		if isDigit(s[0]) {
			return true
		}
		return string(s) < string(t)
	}
	sps := string(s[nprefix:])
	tps := string(t[nprefix:])
	if strings.HasPrefix(sps, "-") || strings.HasPrefix(tps, "-") {
		// digits < -
		if leadDigits(s[nprefix:]) > 0 && strings.HasPrefix(tps, "-") {
			fmt.Println("HERE1")
			return true
		}
		if leadDigits(t[nprefix:]) > 0 && strings.HasPrefix(sps, "-") {
			fmt.Println("HERE2")
			return false
		}
		// . < -
		if strings.HasPrefix(sps, ".") && strings.HasPrefix(tps, "-") {
			fmt.Println("HERE5")
			return false
		}
		if strings.HasPrefix(sps, "-") && strings.HasPrefix(tps, ".") {
			fmt.Println("HERE6")
			return true
		}
	}
	// digits < .
	if leadDigits(s[nprefix:]) > 0 && strings.HasPrefix(tps, ".") {
		fmt.Println("HERE3")
		return true
	}
	if leadDigits(t[nprefix:]) > 0 && strings.HasPrefix(sps, ".") {
		fmt.Println("HERE4")
		return false
	}
	sEnd := leadDigits(s[nprefix:]) + nprefix
	tEnd := leadDigits(t[nprefix:]) + nprefix
	if sEnd > nprefix || tEnd > nprefix {
		start := trailDigits(s[:nprefix])
		if sEnd-start > 0 && tEnd-start > 0 {
			// TODO(light): log errors?
			sn := atoi(s[start:sEnd])
			tn := atoi(t[start:tEnd])
			if sn != tn {
				return sn < tn
			}
		}
	}
	switch {
	case len(s) == nprefix:
		return true
	case len(t) == nprefix:
		return false
	default:
		return s[nprefix] < t[nprefix]
	}
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func atoi(r []rune) uint64 {
	if len(r) < 1 {
		panic(errors.New("atoi got an empty slice"))
	}
	const cutoff = uint64((1<<64-1)/10 + 1)
	const maxVal = 1<<64 - 1

	var n uint64
	for _, d := range r {
		v := uint64(d - '0')
		if n >= cutoff {
			return 1<<64 - 1
		}
		n *= 10
		n1 := n + v
		if n1 < n || n1 > maxVal {
			// n+v overflows
			return 1<<64 - 1
		}
		n = n1
	}
	return n
}

func commonPrefix(s, t []rune) int {
	for i := range s {
		if i >= len(t) {
			return len(t)
		}
		if s[i] != t[i] {
			return i
		}
	}
	return len(s)
}

func trailDigits(r []rune) int {
	for i := len(r) - 1; i >= 0; i-- {
		if !isDigit(r[i]) {
			return i + 1
		}
	}
	return 0
}

func leadDigits(r []rune) int {
	for i := range r {
		if !isDigit(r[i]) {
			return i
		}
	}
	return len(r)
}

type stringSlice []string

func (ss stringSlice) Len() int {
	return len(ss)
}

func (ss stringSlice) Less(i, j int) bool {
	return Less(ss[i], ss[j])
}

func (ss stringSlice) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
