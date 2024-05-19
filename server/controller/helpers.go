package controller

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
)

func (c *Controller) mustGetTimezone(ctx echo.Context) *time.Location {
	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		panic(err)
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		panic(err)
	}

	return timezone
}

func (c *Controller) midnightInLocal(ctx echo.Context, input time.Time) (time.Time, error) {
	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return input, errors.Wrap(err, "failed to retrieve account's timezone")
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		return input, errors.Wrap(err, "failed to parse account's timezone")
	}

	return util.Midnight(input, timezone), nil
}

func (c *Controller) getLoginId(ctx echo.Context) (ID[Login], error) {
	loginId, ok := ctx.Get(loginIdContextKey).(string)
	if loginId == "" || !ok {
		return "", errors.New("unauthorized: no loginId present on request")
	}

	parsed, err := ParseID[Login](loginId)
	if err != nil {
		return "", errors.Wrap(err, "unauthorized: loginId on request is invalid")
	}

	return parsed, nil
}

// getUserId will take the current request context and will look for a user ID
// on the context. If one is present it will be returned. If there is not one
// present or if it is not in the correct format then an error will be returned
// and a "zero" ID will be returned instead.
func (c *Controller) getUserId(ctx echo.Context) (ID[User], error) {
	userId, ok := ctx.Get(userIdContextKey).(string)
	if userId == "" || !ok {
		return "", errors.New("unauthorized: no userId present on request")
	}

	parsed, err := ParseID[User](userId)
	if err != nil {
		return "", errors.Wrap(err, "unauthorized: userId on request is invalid")
	}

	return parsed, nil
}

func (c *Controller) mustGetUserId(ctx echo.Context) ID[User] {
	userId, err := c.getUserId(ctx)
	if err != nil {
		panic(err)
	}

	return userId
}

func (c *Controller) getAccountId(ctx echo.Context) (ID[Account], error) {
	accountId, ok := ctx.Get(accountIdContextKey).(string)
	if accountId == "" || !ok {
		return "", errors.New("unauthorized: no accountId present on request")
	}

	parsed, err := ParseID[Account](accountId)
	if err != nil {
		return "", errors.Wrap(err, "unauthorized: accountId on request is invalid")
	}

	return parsed, nil
}

func (c *Controller) mustGetAccountId(ctx echo.Context) ID[Account] {
	accountId, err := c.getAccountId(ctx)
	if err != nil {
		panic(err)
	}

	return accountId
}

func (c *Controller) mustGetDatabase(ctx echo.Context) pg.DBI {
	txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
	if !ok {
		panic("no database on context")
	}

	return txn
}

// mustGetSecurityRepository is used to retrieve/create a repository interface
// that can interact with more security sensitive parts of the data layer. This
// interface is not specific to a single tenant. If the interface cannot be
// created due then this method will panic.
func (c *Controller) mustGetSecurityRepository(ctx echo.Context) repository.SecurityRepository {
	db, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		panic("failed to retrieve database object from controller context")
	}

	return repository.NewSecurityRepository(db, c.clock)
}

func (c *Controller) getUnauthenticatedRepository(ctx echo.Context) (repository.UnauthenticatedRepository, error) {
	txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(c.clock, txn), nil
}

func (c *Controller) mustGetUnauthenticatedRepository(ctx echo.Context) repository.UnauthenticatedRepository {
	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		panic(err)
	}

	return repo
}

func (c *Controller) getAuthenticatedRepository(ctx echo.Context) (repository.Repository, error) {
	userId, err := c.getUserId(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "you are not authenticated to an account")
	}

	accountId, err := c.getAccountId(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "you are not authenticated to an account")
	}

	txn, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewRepositoryFromSession(c.clock, userId, accountId, txn), nil
}

func (c *Controller) mustGetAuthenticatedRepository(ctx echo.Context) repository.Repository {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		panic("unauthorized")
	}

	return repo
}

func (c *Controller) getSecretsRepository(ctx echo.Context) (repository.SecretsRepository, error) {
	accountId, err := c.getAccountId(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "you are not authenticated to an account")
	}

	txn, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	log := c.getLog(ctx)

	return repository.NewSecretsRepository(
		log,
		c.clock,
		txn,
		c.kms,
		accountId,
	), nil
}

func (c *Controller) mustGetSecretsRepository(ctx echo.Context) repository.SecretsRepository {
	repo, err := c.getSecretsRepository(ctx)
	if err != nil {
		panic("unauthorized")
	}

	return repo
}
