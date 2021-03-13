package controller

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
)

const (
	transactionContextKey = "_harderTransaction_"
	accountIdContextKey   = "_accountId_"
	userIdContextKey      = "_userId_"
	loginIdContextKey     = "_loginId_"
)

func (c *Controller) setupRepositoryMiddleware(ctx *context.Context) {
	txn, err := c.db.BeginContext(ctx.Request().Context())
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to begin transaction")
		return
	}

	ctx.Values().Set(transactionContextKey, txn)

	ctx.Next()

	// TODO (elliotcourant) Add proper logging here that the request has failed
	//  and we are rolling back the transaction.
	if ctx.GetErr() != nil {
		if err := txn.RollbackContext(ctx.Request().Context()); err != nil {
			// Rollback
			c.log.WithError(err).Errorf("failed to rollback request")
		}
	} else {
		if err = txn.CommitContext(ctx.Request().Context()); err != nil {
			// failed to commit
			fmt.Println(err)
		}
	}

	ctx.Next()
}

func (c *Controller) authenticationMiddleware(ctx *context.Context) {
	token := ctx.GetHeader(TokenName)
	if token == "" {
		c.returnError(ctx, http.StatusForbidden, "unauthorized")
		return
	}

	var claims HarderClaims
	result, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(c.configuration.JWT.LoginJwtSecret), nil
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "unauthorized")
		return
	}

	if !result.Valid {
		c.returnError(ctx, http.StatusForbidden, "unauthorized")
		return
	}

	ctx.Values().Set(accountIdContextKey, claims.AccountId)
	ctx.Values().Set(userIdContextKey, claims.UserId)
	ctx.Values().Set(loginIdContextKey, claims.LoginId)

	ctx.Next()
}

func (c *Controller) loggingMiddleware(ctx *context.Context) {
	ctx.Next()

	if err := ctx.GetErr(); err != nil {
		c.log.WithContext(ctx.Request().Context()).WithError(err).Errorf("%+v", ctx.GetErr())
	}
}

func (c *Controller) getUnauthenticatedRepository(ctx *context.Context) (repository.UnauthenticatedRepository, error) {
	txn, ok := ctx.Values().Get(transactionContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(txn), nil
}

func (c *Controller) mustGetUserId(ctx *context.Context) uint64 {
	userId := ctx.Values().GetUint64Default(userIdContextKey, 0)
	if userId == 0 {
		panic("unauthorized")
	}

	return userId
}

func (c *Controller) mustGetAccountId(ctx *context.Context) uint64 {
	accountId := ctx.Values().GetUint64Default(accountIdContextKey, 0)
	if accountId == 0 {
		panic("unauthorized")
	}

	return accountId
}

func (c *Controller) getAuthenticatedRepository(ctx *context.Context) (repository.Repository, error) {
	loginId := ctx.Values().GetUint64Default(loginIdContextKey, 0)
	if loginId == 0 {
		return nil, errors.Errorf("not authorized")
	}

	userId := ctx.Values().GetUint64Default(userIdContextKey, 0)
	if userId == 0 {
		return nil, errors.Errorf("you are not authenticated to an account")
	}

	accountId := ctx.Values().GetUint64Default(accountIdContextKey, 0)
	if accountId == 0 {
		return nil, errors.Errorf("you are not authenticated to an account")
	}

	txn, ok := ctx.Values().Get(transactionContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewRepositoryFromSession(userId, accountId, txn), nil
}

func (c *Controller) mustGetAuthenticatedRepository(ctx *context.Context) repository.Repository {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		panic("unauthorized")
	}

	return repo
}
