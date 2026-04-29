package myownsanity

import "errors"

func FirstError(items ...error) error {
	for _, item := range items {
		if item != nil {
			return item
		}
	}
	return nil
}

// JoinErrorMaybe takes an array of errors, if any of the errors are nil then
// nil is returned. If all of the errors are actually errors then the joined
// version of the errors is returned.
func JoinErrorMaybe(items ...error) error {
	for _, item := range items {
		if item == nil {
			return nil
		}
	}

	return errors.Join(items...)
}
