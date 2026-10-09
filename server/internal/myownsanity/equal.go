package myownsanity

// PEqual will compare whether or not two pointers point to equal values. If
// only one of them is nil then it will return false, if both are nil then they
// are equal.
func PEqual[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}
