package controller

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
)

const (
	transactionContextKey = "_harderTransaction_"
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
			fmt.Println(err)
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

}

func (c *Controller) getUnauthenticatedRepository(ctx *context.Context) (repository.UnauthenticatedRepository, error) {
	txn, ok := ctx.Values().Get(transactionContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(txn), nil
}
