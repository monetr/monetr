package middleware

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/security"
)

const (
	APIKeyHeader = "X-API-Key"
	authenticationKey = "_authentication_"
)

func APIKeyAuthentication(repo repository.APIKeyRepository, db pg.DBI) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if already authenticated via session
			if c.Get("userId") != nil {
				return next(c)
			}

			apiKey := c.Request().Header.Get(APIKeyHeader)
			if apiKey == "" {
				return next(c)
			}

			// Hash the provided key
			hash := sha256.Sum256([]byte(apiKey))
			keyHash := base64.URLEncoding.EncodeToString(hash[:])

			// Look up the API key
			key, err := repo.GetAPIKeyByHash(c.Request().Context(), keyHash)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}

			// Check if key is expired
			if !key.ExpiresAt.IsZero() && key.ExpiresAt.Before(time.Now()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "API key expired")
			}

			// Update last used timestamp
			go repo.UpdateAPIKeyLastUsed(c.Request().Context(), key.APIKeyId)

			// Look up the user to get the account ID
			var user models.User
			err = db.ModelContext(c.Request().Context(), &user).
				Where("user_id = ?", key.UserId).
				Select()
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not found for API key")
			}

			// Create and set security claims for the API key
			claims := security.Claims{
				CreatedAt: time.Now(),
				UserId:    key.UserId,
				LoginId:   string(user.LoginId),
				AccountId: string(user.AccountId),
				Scope:     security.AuthenticatedScope,
			}
			
			// Store the authentication claims on the request context
			c.Set(authenticationKey, claims)
			
			// Set the user ID in context
			c.Set("userId", key.UserId)
			
			return next(c)
		}
	}
}
