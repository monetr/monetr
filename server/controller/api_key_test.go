package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
)

// givenIHaveTokenWithProofOfWork registers and logs in a fresh user against a
// proof of work enabled server, solving the register and login challenges along
// the way, and returns the resulting session token.
func givenIHaveTokenWithProofOfWork(t *testing.T, e *httpexpect.Expect) string {
	registerBody := validRegisterBody(t)
	{
		challenge, nonce := getAndSolveChallenge(t, e, "register")
		registerBody["challenge"] = challenge
		registerBody["nonce"] = nonce
		e.POST("/api/authentication/register").
			WithJSON(registerBody).
			Expect().
			Status(http.StatusOK)
	}

	challenge, nonce := getAndSolveChallenge(t, e, "login")
	response := e.POST("/api/authentication/login").
		WithJSON(map[string]any{
			"email":     registerBody["email"],
			"password":  registerBody["password"],
			"challenge": challenge,
			"nonce":     nonce,
		}).
		Expect()
	response.Status(http.StatusOK)
	return response.Cookie(TestCookieName).Value().NotEmpty().Raw()
}

// createApiKeyWithProofOfWork creates an API key against a proof of work enabled
// server and returns the new key's Id.
func createApiKeyWithProofOfWork(t *testing.T, e *httpexpect.Expect, token string) string {
	challenge, nonce := getAndSolveChallenge(t, e, "create_api_key")
	return e.POST("/api/keys").
		WithCookie(TestCookieName, token).
		WithJSON(map[string]any{
			"name":      gofakeit.UUID(),
			"challenge": challenge,
			"nonce":     nonce,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("apiKeyId").String().NotEmpty().Raw()
}

func TestPostApiKey(t *testing.T) {
	t.Run("proof of work disabled", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "My First Key",
			}).
			Expect()
		response.Status(http.StatusOK)
		response.JSON().Path("$.apiKeyId").String().NotEmpty()
		response.JSON().Path("$.secret").String().NotEmpty()
		response.JSON().Path("$.name").String().IsEqual("My First Key")
	})

	t.Run("proof of work required when enabled", func(t *testing.T) {
		_, e := NewTestApplicationWithConfig(t, powEnabledConfig(t))
		token := givenIHaveTokenWithProofOfWork(t, e)

		// Without a solved challenge the request is rejected by the schema.
		response := e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "My First Key",
			}).
			Expect()
		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
	})

	t.Run("proof of work accepted when enabled", func(t *testing.T) {
		_, e := NewTestApplicationWithConfig(t, powEnabledConfig(t))
		token := givenIHaveTokenWithProofOfWork(t, e)

		challenge, nonce := getAndSolveChallenge(t, e, "create_api_key")
		response := e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "My First Key",
				"challenge": challenge,
				"nonce":     nonce,
			}).
			Expect()
		response.Status(http.StatusOK)
		response.JSON().Path("$.secret").String().NotEmpty()
	})

	t.Run("rejects a challenge issued for another purpose", func(t *testing.T) {
		_, e := NewTestApplicationWithConfig(t, powEnabledConfig(t))
		token := givenIHaveTokenWithProofOfWork(t, e)

		// A challenge solved for logging in must not be usable to create a key,
		// the purpose is bound into the signed token.
		challenge, nonce := getAndSolveChallenge(t, e, "login")
		response := e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "My First Key",
				"challenge": challenge,
				"nonce":     nonce,
			}).
			Expect()
		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("invalid proof of work")
	})

	t.Run("rejects a replayed challenge", func(t *testing.T) {
		_, e := NewTestApplicationWithConfig(t, powEnabledConfig(t))
		token := givenIHaveTokenWithProofOfWork(t, e)

		challenge, nonce := getAndSolveChallenge(t, e, "create_api_key")
		e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "First Key",
				"challenge": challenge,
				"nonce":     nonce,
			}).
			Expect().
			Status(http.StatusOK)

		// Reusing the same challenge and nonce must be rejected as a replay.
		response := e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name":      "Second Key",
				"challenge": challenge,
				"nonce":     nonce,
			}).
			Expect()
		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("challenge already used")
	})
}

func TestDeleteApiKey(t *testing.T) {
	t.Run("proof of work disabled", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		keyId := e.POST("/api/keys").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "To Be Deleted",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object().Value("apiKeyId").String().NotEmpty().Raw()

		// No body is required when proof of work is disabled.
		e.DELETE("/api/keys/{apiKeyId}").
			WithPath("apiKeyId", keyId).
			WithCookie(TestCookieName, token).
			Expect().
			Status(http.StatusOK)
	})

	t.Run("proof of work required when enabled", func(t *testing.T) {
		_, e := NewTestApplicationWithConfig(t, powEnabledConfig(t))
		token := givenIHaveTokenWithProofOfWork(t, e)
		keyId := createApiKeyWithProofOfWork(t, e, token)

		// A delete with no challenge body must be rejected.
		e.DELETE("/api/keys/{apiKeyId}").
			WithPath("apiKeyId", keyId).
			WithCookie(TestCookieName, token).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("proof of work accepted when enabled", func(t *testing.T) {
		_, e := NewTestApplicationWithConfig(t, powEnabledConfig(t))
		token := givenIHaveTokenWithProofOfWork(t, e)
		keyId := createApiKeyWithProofOfWork(t, e, token)

		challenge, nonce := getAndSolveChallenge(t, e, "delete_api_key")
		e.DELETE("/api/keys/{apiKeyId}").
			WithPath("apiKeyId", keyId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"challenge": challenge,
				"nonce":     nonce,
			}).
			Expect().
			Status(http.StatusOK)
	})
}
