package myownsanity

import (
	"fmt"
)

func ASSERT_NOTNIL(value any, msg string) {
	if value == nil {
		panic(fmt.Sprintf("assert failed: %+v is nil, %s", value, msg))
	}
}

// MUST is a function that wraps another function that must succeed. This
// generally should not be used in most code paths and should instead be used
// inside of init() functions or other entrypoints such that the problem becomes
// evident immediately upon starting or testing monetr as opposed to some
// obscure codepath that must be hit first.
func MUST[T any, A any](callback func(arg A) (T, error), arg A) T {
	result, err := callback(arg)
	if err != nil {
		panic(fmt.Sprintf("MUST FAILED!\n%+v", err))
	}
	return result
}
