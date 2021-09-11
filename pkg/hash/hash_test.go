package hash

import (
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
)

func TestHashEmail(t *testing.T) {
	hash := HashEmail("me@elliotcourant.dev")
	fmt.Println(hash)

	testEmails := func(numberOfEmailsToTest int) func(t *testing.T) {
		return func(t *testing.T) {
			collisions := 0
			uniqueEmails := map[string]string{}
			hashedEmails := map[string]string{}

			for len(uniqueEmails) < numberOfEmailsToTest {
				email := strings.ToLower(gofakeit.Email())
				if _, ok := uniqueEmails[email]; ok {
					// Keep track of how many times gofakeit generates a duplicate email.
					collisions++
					// If we get too many then panic.
					if collisions > 1000 {
						panic("too many collisions")
					}
					continue
				}

				hashed := HashEmail(email)
				assert.NotEmptyf(t, hashed, "hashed email must not be empty, input: %s", email)
				assert.Equalf(t, hashed, HashEmail(email), "must match hash consistently, input: %s", email)
				if !assert.NotContainsf(t, hashedEmails, hashed, "hashed email must not be duplicate, input: %s", email) {
					fmt.Println("two emails produced the same hash, a:", email, "b:", hashedEmails[hashed], "hash:", hashed)
					return
				}

				uniqueEmails[email] = hashed
				hashedEmails[hashed] = email
			}

			assert.Equal(t, len(uniqueEmails), len(hashedEmails), "number of unique and hashed emails should match")
		}
	}

	t.Run("100", testEmails(100))
	t.Run("1000", testEmails(1000))
	t.Run("10000", testEmails(10000))
}
