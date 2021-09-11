package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/OneOfOne/xxhash"
)

// HashEmail will take an email address as an input and output a hash that is unique to that email address. This can be
// used to uniquely identify email addresses without needing to store the address itself.
func HashEmail(email string) string {
	// Cast the email to be lower case, this way we can store the email in a way that is "case-insensitive". Since every
	// email will be hashed and checksum-ed as lower case.
	email = strings.ToLower(email)
	// Create a SHA256 hash of the now lower case email address.
	hash := sha256.New()
	hash.Write([]byte(email))
	// Make a byte array with the result.
	result := hash.Sum(nil)
	// Add 8 bytes on the end for a 64 bit checksum.
	result = append(result, make([]byte, 8)...)
	// Create a checksum of the un-hashed email address.
	checksum := xxhash.ChecksumString64S(email, uint64(len(email)))
	// Write that checksum to the 8 bytes at the end.
	binary.BigEndian.PutUint64(result[len(result)-8:], checksum)
	// Return the hexadecimal representation of the hash+checksum-ed email.
	return fmt.Sprintf("%X", result)
}

// HashPassword will return a one way hash of the provided user's credentials.
// The email is always converted to lowercase for this hash but the password is
// not modified.
func HashPassword(email, password string) string {
	email = strings.ToLower(email)
	hash := sha256.New()
	hash.Write([]byte(email))
	hash.Write([]byte(password))
	return fmt.Sprintf("%X", hash.Sum(nil))
}
