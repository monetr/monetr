package controller

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

// cleanString takes the current request context, the name of a field and the
// user input for that field. It will validate that the input is not longer than
// 250 characters and will return a bad request error if it is. It will also
// return a whitespace trimmed version of the provided input.
func (c *Controller) cleanString(ctx echo.Context, name string, input string) (string, error) {
	if len(input) > 250 {
		return input, c.badRequest(ctx, "%s must not be longer than 250 characters", name)
	}
	// Trim after asserting length. This way if they send an incredibly long
	// string we are not wasting time only to return an error.
	input = strings.TrimSpace(input)
	return input, nil
}

func (c *Controller) mustGetTimezone(ctx echo.Context) *time.Location {
	account, err := c.Accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
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
	account, err := c.Accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return input, errors.Wrap(err, "failed to retrieve account's timezone")
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		return input, errors.Wrap(err, "failed to parse account's timezone")
	}

	return util.Midnight(input, timezone), nil
}

func (c *Controller) getClaims(ctx echo.Context) (security.Claims, error) {
	claims, ok := ctx.Get(authenticationKey).(security.Claims)
	if !ok {
		return claims, errors.New("unauthorized: claims not present on request")
	}

	return claims, claims.Valid()
}

func (c *Controller) mustGetClaims(ctx echo.Context) security.Claims {
	claims, err := c.getClaims(ctx)
	if err != nil {
		panic("unauthorized: claims on request are invalid")
	}

	return claims
}

func (c *Controller) getLoginId(ctx echo.Context) (ID[Login], error) {
	claims, err := c.getClaims(ctx)
	if err != nil {
		return "", err
	}

	parsed, err := ParseID[Login](claims.LoginId)
	if err != nil {
		return "", errors.Wrap(err, "unauthorized: loginId on request is invalid")
	}

	return parsed, nil
}

func (c *Controller) mustGetLoginId(ctx echo.Context) ID[Login] {
	loginId, err := c.getLoginId(ctx)
	if err != nil {
		panic(err)
	}

	return loginId
}

// getUserId will take the current request context and will look for a user ID
// on the context. If one is present it will be returned. If there is not one
// present or if it is not in the correct format then an error will be returned
// and a "zero" ID will be returned instead.
func (c *Controller) getUserId(ctx echo.Context) (ID[User], error) {
	claims, err := c.getClaims(ctx)
	if err != nil {
		return "", err
	}

	if claims.UserId == "" {
		return "", errors.New("unauthorized: no userId present on request")
	}

	parsed, err := ParseID[User](claims.UserId)
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
	claims, err := c.getClaims(ctx)
	if err != nil {
		return "", err
	}

	if claims.AccountId == "" {
		return "", errors.New("unauthorized: no accountId present on request")
	}

	parsed, err := ParseID[Account](claims.AccountId)
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

	return repository.NewSecurityRepository(db, c.Clock)
}

func (c *Controller) getUnauthenticatedRepository(ctx echo.Context) (repository.UnauthenticatedRepository, error) {
	txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(c.Clock, txn), nil
}

func (c *Controller) mustGetUnauthenticatedRepository(ctx echo.Context) repository.UnauthenticatedRepository {
	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		panic(err)
	}

	return repo
}

func (c *Controller) getAuthenticatedRepository(
	ctx echo.Context,
) (repository.Repository, error) {
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

	return repository.NewRepositoryFromSession(
		c.Clock,
		userId,
		accountId,
		txn,
		c.getLog(ctx),
	), nil
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
		c.Clock,
		txn,
		c.KMS,
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

func (c *Controller) scrubSentryBody(ctx echo.Context) {
	// If sentry is setup, make sure we never send the body for this request to
	// sentry.
	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		if scope := hub.Scope(); scope != nil {
			scope.SetRequestBody(nil)
		}
	}
}

func (c *Controller) sendWebsocketMessage(ctx echo.Context, ws *websocket.Conn, message any) error {
	log := c.getLog(ctx)
	msg, err := json.Marshal(message)
	if err != nil {
		log.WithField("mesasge", message).WithError(err).Error("failed to encode websocket message")
		return err
	}
	if err := websocket.Message.Send(ws, string(msg)); err != nil {
		log.WithField("mesasge", message).WithError(err).Error("failed to send websocket message")
		return err
	}

	return nil
}

func (c *Controller) readWebsocketMessage(ctx echo.Context, ws *websocket.Conn, result any) error {
	log := c.getLog(ctx)

	var message string
	if err := websocket.Message.Receive(ws, &message); err != nil {
		log.WithField("message", message).WithError(err).Error("failed to read websocket message")
		return errors.Wrap(err, "failed to read websocket message")
	}

	return errors.Wrap(json.Unmarshal([]byte(message), result), "failed to decode websocket message")
}
