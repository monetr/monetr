package myownsanity

func Map[A any, B any, F func(arg A) B](items []A, callback F) []B {
	result := make([]B, 0, len(items))
	for _, item := range items {
		result = append(result, callback(item))
	}

	return result
}
