package myownsanity

// Union returns the distinct items in both a and b in order from a to b.
// https://en.wikipedia.org/wiki/Set_(mathematics)#Union
func Union[T comparable](a []T, b []T) []T {
	unique := map[T]struct{}{}
	accumulator := make([]T, 0, len(a)+len(b))
	for _, items := range [][]T{a, b} {
		for _, item := range items {
			if _, ok := unique[item]; ok {
				continue
			}
			unique[item] = struct{}{}
			accumulator = append(accumulator, item)
		}
	}
	return accumulator
}
