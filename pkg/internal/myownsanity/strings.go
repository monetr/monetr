package myownsanity

func StringP(input string) *string {
	return &input
}

func SliceContains(slice []string, item string) bool {
	for _, sliceItem := range slice {
		if item == sliceItem {
			return true
		}
	}

	return false
}
