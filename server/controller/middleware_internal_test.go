package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseRepositoryMiddleware(t *testing.T) {
	t.Run("cancelled request rolls back without panicking", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)
		c := &Controller{DB: db, Log: log}

		requestCtx, cancel := context.WithCancel(t.Context())
		defer cancel()
		req := httptest.NewRequest(http.MethodPost, "/test", nil).WithContext(requestCtx)
		ctx := echo.New().NewContext(req, httptest.NewRecorder())

		account := models.Account{
			Timezone: "UTC",
			Locale:   "en_US",
		}
		handler := c.databaseRepositoryMiddleware(func(ctx *echo.Context) error {
			dbi := c.mustGetDatabase(ctx)
			_, err := dbi.NewInsert().Model(&account).Exec(c.getContext(ctx))
			require.NoError(t, err, "must be able to insert inside the request transaction")

			// The client goes away after the handler has done its writes but before
			// the middleware commits the transaction.
			cancel()
			return nil
		})

		assert.NotPanics(t, func() {
			assert.NoError(t, handler(ctx), "handler should not return an error")
		}, "cancelled request must not panic on commit")

		exists, err := db.NewSelect().
			Model((*models.Account)(nil)).
			Where(`"account"."account_id" = ?`, account.AccountId).
			Exists(context.Background())
		require.NoError(t, err, "must be able to check for the account")
		assert.False(t, exists, "write from the cancelled request should have been rolled back")

		testutils.MustHaveLogMessage(t, hook, "request cancelled, transaction was rolled back")
	})
}
