package myownsanity

// Intersection returns the unique items that exist in both of the provided
// arrays [a] and [b] in the order that they are observed from the shorter of
// the two arrays.
// https://en.wikipedia.org/wiki/Set_(mathematics)#Intersection
func Intersection[T comparable](a []T, b []T) []T {
	// basis is the shorter of the two arrays
	basis := a
	axis := b
	if len(b) < len(basis) {
		basis = b
		axis = a
	}

	// The intersection will not be longer than the shorter of the two arrays.
	intersection := make([]T, 0, len(basis))
	unique := map[T]struct{}{}

Loop:
	for _, needle := range basis {
		// If the basis array has duplicate values, skip them. The intersection
		// returned will be unique.
		if _, ok := unique[needle]; ok {
			continue Loop
		}
		for _, item := range axis {
			if needle == item {
				intersection = append(intersection, item)
				unique[needle] = struct{}{}
				// We don't need to continue looking on the inner loop because it is not
				// valuable for us to find another one of the same needle.
				continue Loop
			}
		}
	}

	return intersection
}
