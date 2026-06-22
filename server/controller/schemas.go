package controller

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/schemas"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
)

func parse[T any](
	c *Controller,
	ctx *echo.Context,
	input *T,
	schema validation.Rule,
) (*T, error) {
	result, err := schemas.Parse(
		c.getContext(ctx),
		ctx.Request().Body,
		input,
		schema,
	)
	switch err := err.(type) {
	case validation.Errors, validation.OneOfError:
		return nil, &apiResponseError{
			code: http.StatusBadRequest,
			body: map[string]any{
				"error":    "Invalid request",
				"problems": validators.MarshalErrorTree(err),
			},
			err: err,
		}
	case *json.SyntaxError:
		return nil, c.invalidJsonError(ctx, err)
	case nil:
		return result, nil
	default:
		return nil, c.badRequestError(ctx, err, "failed to parse request")
	}
}
