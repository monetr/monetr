package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/datasources/table"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

func (c *Controller) getTransactionImportMappings(ctx echo.Context) error {
	limit := urlParamIntDefault(ctx, "limit", 10)
	offset := urlParamIntDefault(ctx, "offset", 0)

	if limit < 1 {
		return c.badRequest(ctx, "limit must be at least 1")
	} else if limit > 10 {
		return c.badRequest(ctx, "limit cannot be greater than 10")
	}

	if offset < 0 {
		return c.badRequest(ctx, "offset cannot be less than 0")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	var mappings []TransactionImportMapping
	var err error
	if signature := ctx.QueryParam("signature"); signature != "" {
		mappings, err = repo.GetTransactionImportMappingsBySignature(
			c.getContext(ctx),
			signature,
			limit,
			offset,
		)
	} else {
		mappings, err = repo.GetTransactionImportMappings(
			c.getContext(ctx),
			limit,
			offset,
		)
	}
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transaction import mappings")
	}

	return ctx.JSON(http.StatusOK, mappings)
}

func (c *Controller) postTransactionImportMapping(ctx echo.Context) error {
	var data struct {
		Mapping table.Mapping `json:"mapping"`
	}
	if err := ctx.Bind(&data); err != nil {
		return c.invalidJson(ctx)
	}

	if err := data.Mapping.Validate(c.getContext(ctx)); err != nil {
		switch errors.Cause(err).(type) {
		case validation.Errors, validators.OneOfError:
			return c.jsonError(ctx, http.StatusBadRequest, map[string]any{
				"error":    "Invalid request",
				"problems": validators.MarshalErrorTree(err),
			})
		default:
			return c.badRequestError(ctx, err, "failed to parse request")
		}
	}

	mapping := TransactionImportMapping{
		Mapping: data.Mapping,
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateTransactionImportMapping(c.getContext(ctx), &mapping); err != nil {
		return c.wrapPgError(ctx, err, "failed to create transaction import mapping")
	}

	return ctx.JSON(http.StatusOK, mapping)
}
