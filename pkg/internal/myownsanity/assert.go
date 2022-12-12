package myownsanity

import (
	"fmt"
)

func Assert(predicate bool, message string) {
	if !predicate {
		panic(fmt.Sprintf("ASSERT FAILED: %s", message))
	}
}
