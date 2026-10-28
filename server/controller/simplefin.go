package controller

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/datasources/simplefin"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/schemas"
	"github.com/pkg/errors"
)

func (c *Controller) getSimpleFinLinks(ctx *echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)
	links, err := repo.GetSimpleFinLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN links")
	}

	return ctx.JSON(http.StatusOK, links)
}

func (c *Controller) getSimpleFinLink(ctx *echo.Context) error {
	id, err := ParseID[SimpleFinLink](ctx.Param("simpleFinLinkId"))
	if err != nil || id.IsZero() {
		return c.badRequest(ctx, "Must specify a valid SimpleFIN Link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetSimpleFinLink(c.getContext(ctx), id)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN link")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) postSimpleFinLink(ctx *echo.Context) error {
	request, err := parse(
		c,
		ctx,
		new(schemas.PostSimpleFinLinkRequest),
		schemas.PostSimpleFinLink,
	)
	if err != nil {
		return err
	}

	accessUrl, err := simplefin.ClaimSetupToken(
		c.getContext(ctx),
		c.getLog(ctx),
		request.SetupToken,
		c.Configuration.SimpleFin,
	)
	switch errors.Cause(err) {
	case nil:
		// Do nothing
	case simplefin.ErrInvalidSetupToken:
		return c.badRequest(ctx, "Invalid SimpleFIN setup token")
	case simplefin.ErrHostNotAllowed:
		return c.badRequest(ctx, "SimpleFIN setup token is for a server that is not allowed")
	case simplefin.ErrSetupTokenClaimed:
		return c.badRequest(ctx, "SimpleFIN setup token has already been used")
	default:
		// Don't send err.Error() back here, if the claim failed then it can have
		// the raw response body from SimpleFIN in it.
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to claim SimpleFIN setup token",
		)
	}

	// The setup token is spent at this point. If anything below fails then the
	// transaction gets rolled back and the access URL is gone, so the user will
	// need to make a new setup token.

	// Keep a copy of the URL without the credentials so we can show it to the
	// user. The client already made sure this parses.
	apiUrl, err := url.Parse(accessUrl)
	if err != nil {
		// Don't include the parse error, it would have the credentials in it.
		return c.returnError(
			ctx,
			http.StatusInternalServerError,
			"Failed to parse SimpleFIN access URL, you will need to create a new setup token",
		)
	}
	apiUrl.User = nil

	secret := repository.SecretData{
		Kind:  SecretKindSimpleFin,
		Value: accessUrl,
	}

	{ // Store the secret and generate an ID
		secrets := c.mustGetSecretsRepository(ctx)
		if err := secrets.Store(c.getContext(ctx), &secret); err != nil {
			return c.wrapPgError(ctx, err, "Failed to store SimpleFIN link, you will need to create a new setup token")
		}
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	simpleFinLink := SimpleFinLink{
		Name:     request.Name,
		SecretId: secret.SecretId,
		ApiUrl:   apiUrl.String(),
		Status:   SimpleFinLinkStatusPending,
	}

	if err := repo.CreateSimpleFinLink(
		c.getContext(ctx),
		&simpleFinLink,
	); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create SimpleFIN link, you will need to create a new setup token")
	}

	return ctx.JSON(http.StatusOK, simpleFinLink)
}
