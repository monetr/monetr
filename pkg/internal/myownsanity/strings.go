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

// StringPEqual will compare whether or not two string pointers are equal. If one or the other is nil then it will
// return false. Otherwise it will compare their values as strings not as pointers.
func StringPEqual(a, b *string) bool {
	if a == nil && b != nil {
		return false
	}
	if a != nil && b == nil {
		return false
	}
	if a == nil && b == nil {
		return true
	}

	return *a == *b
}