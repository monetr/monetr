package myownsanity

func StringP(input string) *string {
	return &input
}

func StringDefault(input *string, defaultValue string) string {
	if input != nil {
		return *input
	}

	return defaultValue
}

func SliceContains(slice []string, item string) bool {
	for _, sliceItem := range slice {
		if item == sliceItem {
			return true
		}
	}

	return false
}
