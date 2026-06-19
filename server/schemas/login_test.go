package schemas_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/powchallenge"
	"github.com/monetr/monetr/server/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginSchema(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"password123"}`
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.NoError(t, err, "a basic email and password should be valid")
		require.NotNil(t, result, "must have parsed the login request")
		assert.Equal(t, "test@example.com", result.Email, "should have parsed the email")
		assert.Equal(t, "password123", result.Password, "should have parsed the password")
	})

	t.Run("email and password are trimmed", func(t *testing.T) {
		// cleanStrings inside Parse trims the string fields before validation even
		// runs so a little stray whitespace should not trip anything up.
		body := `{"email":"  test@example.com  ","password":"  password123  "}`
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.NoError(t, err, "surrounding whitespace should be trimmed and the request should be valid")
		require.NotNil(t, result, "must have parsed the login request")
		assert.Equal(t, "test@example.com", result.Email, "the email should have been trimmed")
		assert.Equal(t, "password123", result.Password, "the password should have been trimmed")
	})

	t.Run("email is required", func(t *testing.T) {
		body := `{"password":"password123"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.EqualError(t, err, "email: required key is missing.", "an email is required to login")
	})

	t.Run("email must be valid", func(t *testing.T) {
		body := `{"email":"not-an-email","password":"password123"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.EqualError(t, err, "email: Email address is not valid.", "a malformed email should be rejected")
	})

	t.Run("email must be lower case", func(t *testing.T) {
		// cleanStrings trims but it does not lowercase, so an email with capital
		// letters survives all the way to validation and should be rejected. The
		// caller is expected to send us a normalized lower case address.
		body := `{"email":"Test@Example.com","password":"password123"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.EqualError(t, err, "email: Email address must be lower case.", "an email with capital letters should be rejected")
	})

	t.Run("password is required", func(t *testing.T) {
		body := `{"email":"test@example.com"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.EqualError(t, err, "password: required key is missing.", "a password is required to login")
	})

	t.Run("password cannot be too short", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"short"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.EqualError(t, err, "password: Password must be between 8 and 72 characters.", "a password under 8 characters should be rejected")
	})

	t.Run("password cannot be too long", func(t *testing.T) {
		// 72 is the bcrypt limit. Anything longer than that gets silently truncated
		// by the hashing down the line so we reject it up front instead of letting
		// the user think those extra characters are doing anything.
		body := `{"email":"test@example.com","password":"` + strings.Repeat("a", 73) + `"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.EqualError(t, err, "password: Password must be between 8 and 72 characters.", "a password over 72 characters should be rejected")
	})

	t.Run("a password of exactly 72 characters is allowed", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"` + strings.Repeat("a", 72) + `"}`
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.NoError(t, err, "a password at exactly the 72 character limit should be allowed")
		assert.NotNil(t, result, "must have parsed the login request")
	})

	t.Run("the challenge schema fields are not required here", func(t *testing.T) {
		// The plain login schema is what we use when proof of work is disabled, so
		// it should not care about a challenge or nonce being absent.
		body := `{"email":"test@example.com","password":"password123"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginSchema,
		)
		assert.NoError(t, err, "the plain login schema should not require a challenge or nonce")
	})
}

func TestLoginChallengeSchema(t *testing.T) {
	// A real challenge token is always exactly powchallenge.EncodedTokenLength (96)
	// base32 characters. The contents do not matter for these tests, only that it
	// is the right length and a valid base32 alphabet, powchallenge.Verify is what
	// actually validates the token, not the schema.
	validChallenge := strings.Repeat("MZXW6YTB", powchallenge.EncodedTokenLength/8)

	t.Run("happy path", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"password123","challenge":"` + validChallenge + `","nonce":42}`
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.NoError(t, err, "a complete challenge request should be valid")
		require.NotNil(t, result, "must have parsed the login request")
		assert.Equal(t, "test@example.com", result.Email, "should have parsed the email")
		assert.Equal(t, validChallenge, result.Challenge, "should have parsed the challenge")
		assert.EqualValues(t, 42, result.Nonce, "should have parsed the nonce into the struct")
	})

	t.Run("email and password are still validated", func(t *testing.T) {
		// The challenge schema shares the same email and password rules as the plain
		// login schema, so a bad password should be caught here too.
		body := `{"email":"test@example.com","password":"short","challenge":"` + validChallenge + `","nonce":42}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "password: Password must be between 8 and 72 characters.", "the password rules should still apply when there is a challenge")
	})

	t.Run("challenge is required", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"password123","nonce":42}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "challenge: required key is missing.", "a challenge is required when proof of work is enabled")
	})

	t.Run("challenge must be base32", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"password123","challenge":"not valid base32!","nonce":42}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "challenge: Challenge must be valid.", "a challenge that is not base32 should be rejected")
	})

	t.Run("challenge must be the right length", func(t *testing.T) {
		// This is valid base32 but it is not the fixed length of one of our tokens,
		// so the length rule should catch it before we ever try to verify it.
		body := `{"email":"test@example.com","password":"password123","challenge":"MZXW6YTB","nonce":42}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "challenge: Challenge must be valid.", "a base32 challenge of the wrong length should be rejected")
	})

	t.Run("nonce is required", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"password123","challenge":"` + validChallenge + `"}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "nonce: required key is missing.", "a nonce is required when proof of work is enabled")
	})

	t.Run("a nonce of zero is allowed", func(t *testing.T) {
		// A nonce of 0 is a perfectly valid proof of work solution. The solver
		// starts counting at 0 so every so often 0 is the answer, we must not
		// reject it.
		body := `{"email":"test@example.com","password":"password123","challenge":"` + validChallenge + `","nonce":0}`
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.NoError(t, err, "a nonce of zero should be allowed")
		require.NotNil(t, result, "must have parsed the login request")
		assert.EqualValues(t, 0, result.Nonce, "the zero nonce should have parsed into the struct")
	})

	t.Run("nonce cannot be negative", func(t *testing.T) {
		// Zero is fine but a negative nonce is nonsense, it cannot fit in the uint64
		// field. The Nonce rule catches it up front so we get a clean validation
		// error instead of a merge failure further down.
		body := `{"email":"test@example.com","password":"password123","challenge":"` + validChallenge + `","nonce":-5}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "nonce: Challenge nonce must be valid.", "a negative nonce should be rejected")
	})

	t.Run("nonce cannot be fractional", func(t *testing.T) {
		body := `{"email":"test@example.com","password":"password123","challenge":"` + validChallenge + `","nonce":3.5}`
		_, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.EqualError(t, err, "nonce: Challenge nonce must be valid.", "a fractional nonce should be rejected")
	})

	t.Run("a large nonce is allowed", func(t *testing.T) {
		// Proof of work solutions can get large so make sure a big value still maps
		// cleanly into the struct without the merge falling over.
		body := `{"email":"test@example.com","password":"password123","challenge":"` + validChallenge + `","nonce":999999999999}`
		result, err := schemas.Parse[schemas.LoginRequest](
			t.Context(),
			strings.NewReader(body),
			nil,
			schemas.LoginChallengeSchema,
		)
		assert.NoError(t, err, "a large positive nonce should be valid")
		require.NotNil(t, result, "must have parsed the login request")
		assert.EqualValues(t, 999999999999, result.Nonce, "should have parsed the large nonce into the struct")
	})
}
