package myownsanity

func Pointer[T any](input T) *T {
	return &input
}
