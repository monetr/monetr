package myownsanity

import (
	"fmt"
)

func Assert(predicate bool, message string) {
	panic(fmt.Sprintf("ASSERT FAILED: %s", message))
}
