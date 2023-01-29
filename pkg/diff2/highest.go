package diff2

// highest returns the highest valid index for an array. The equiv of len(s)-1, but with
// less likelihood that you'll commit an off-by-one error.
func highest[S ~[]T, T any](s S) int {
	return len(s) - 1
}
