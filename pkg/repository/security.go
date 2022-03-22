package repository

import (
	"context"
)

type SecurityRepository interface {
	// ChangePassword accepts a login ID and the old hashed password and the new hashed password. The two passwords
	// should be hashed from the user's input. Specifically, you should not retrieve the "oldHashedPassword" from the
	// database and then use it as input for this method. This way the function will only succeed if the provided input
	// is 100% valid. This method will return true if the oldHashedPassword is correct and the update succeeds, it will
	// return false if the oldHashedPassword is incorrect and/or if the update fail.
	ChangePassword(ctx context.Context, loginId uint64, oldHashedPassword, newHashedPassword string) (ok bool, _ error)
}
