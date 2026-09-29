package myownsanity

func Map[A any, B any](items []A, callback func(arg A) B) []B {
	result := make([]B, 0, len(items))
	for _, item := range items {
		result = append(result, callback(item))
	}

	return result
}

func MapErr[A any, B any](items []A, callback func(arg A) (B, error)) ([]B, error) {
	result := make([]B, 0, len(items))
	for _, item := range items {
		output, err := callback(item)
		if err != nil {
			return nil, err
		}
		result = append(result, output)
	}

	return result, nil
}
