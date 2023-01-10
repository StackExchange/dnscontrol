package diff2

import "golang.org/x/exp/constraints"

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
