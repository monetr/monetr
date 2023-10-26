package hash

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Deprecated: use bcrypt instead.
// HashPassword will return a one way hash of the provided user's credentials.
// The email is always converted to lowercase for this hash but the password is
// not modified.
// DEPRECATED! Use bcrypt instead.
func HashPassword(email, password string) string {
	email = strings.ToLower(email)
	hash := sha256.New()
	hash.Write([]byte(email))
	hash.Write([]byte(password))
	return fmt.Sprintf("%X", hash.Sum(nil))
}
