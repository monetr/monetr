package myownsanity

import (
	"fmt"
)

func ASSERT_NOTNIL(value any, msg string) {
	if value == nil {
		panic(fmt.Sprintf("assert failed: %+v is nil, %s", value, msg))
	}
}
