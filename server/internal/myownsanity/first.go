package myownsanity

func FirstError(items ...error) error {
	for _, item := range items {
		if item != nil {
			return item
		}
	}
	return nil
}
