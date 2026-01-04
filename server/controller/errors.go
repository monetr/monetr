package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
)

// wrapPgError will wrap and return an error to the client. But will try to infer a status code from the error it is
// given. If it cannot infer a status code, an InternalServerError is used.
func (c *Controller) wrapPgError(ctx echo.Context, err error, msg string, args ...any) error {
	switch errors.Cause(err) {
	case pg.ErrNoRows:
		friendlyError := fmt.Sprintf("%s: record does not exist", fmt.Sprintf(msg, args...))

		crumbs.Error(
			c.getContext(ctx),
			fmt.Sprintf(msg, args...),
			ctx.Request().URL.Hostname(),
			map[string]any{
				"error": friendlyError,
			},
		)

		return echo.NewHTTPError(
			http.StatusNotFound,
			fmt.Sprintf("%s: record does not exist", fmt.Sprintf(msg, args...)),
		).WithInternal(err)
	default:
		switch actualErr := errors.Cause(err).(type) {
		case pg.Error:
			cleanedErr, status := c.sanitizePgError(actualErr)
			switch status {
			case http.StatusInternalServerError:
				// This will make the cleaned error not visible to the client.
				return c.wrapAndReturnError(ctx, cleanedErr, status, msg, args...)
			default:
				formattedMessage := fmt.Sprint(fmt.Sprintf(msg, args...), ": ", cleanedErr.Error())
				return c.wrapAndReturnError(ctx, cleanedErr, status, formattedMessage, []any{}...)
			}
		default:
			if errors.Is(err, context.DeadlineExceeded) {
				return c.wrapAndReturnError(ctx, err, http.StatusRequestTimeout, msg, args...)
			}
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, msg, args...)
		}
	}
}

func (c *Controller) sanitizePgError(err pg.Error) (error, int) {
	switch err.Field(67) {
	case "23505": // Duplicate
		// TODO Return actual duplicate information in this error.
		return errors.New("a similar object already exists"), http.StatusBadRequest
	default:
		return err, http.StatusInternalServerError
	}
}

func (c *Controller) wrapAndReturnError(ctx echo.Context, err error, status int, msg string, args ...any) error {
	wrapped := errors.Wrapf(err, msg, args...)
	switch status {
	case http.StatusInternalServerError:
		c.reportError(ctx, wrapped)
		fallthrough
	default:
		crumbs.Error(
			c.getContext(ctx),
			fmt.Sprintf(msg, args...),
			ctx.Request().URL.Hostname(),
			map[string]any{
				"error": wrapped.Error(),
			},
		)
		return echo.NewHTTPError(status, fmt.Sprintf(msg, args...)).WithInternal(wrapped)
	}
}

func (c *Controller) failure(ctx echo.Context, status int, error GenericAPIError) error {
	crumbs.Error(
		c.getContext(ctx),
		error.FriendlyMessage(),
		ctx.Request().URL.Hostname(),
		map[string]any{
			"error": error,
		},
	)

	return echo.NewHTTPError(status, error.Error()).WithInternal(error)
}

func (c *Controller) returnError(ctx echo.Context, status int, msg string, args ...any) error {
	err := errors.Errorf(msg, args...)

	crumbs.Error(
		c.getContext(ctx),
		fmt.Sprintf(msg, args...),
		ctx.Request().URL.Hostname(),
		map[string]any{
			"error": err.Error(),
		},
	)

	return echo.NewHTTPError(status, fmt.Sprintf(msg, args...)).WithInternal(err)
}

func (c *Controller) unauthorized(ctx echo.Context) error {
	c.getSpan(ctx).Status = sentry.SpanStatusUnauthenticated
	c.updateAuthenticationCookie(ctx, ClearAuthentication)
	return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
}

func (c *Controller) unauthorizedError(ctx echo.Context, err error) error {
	c.getSpan(ctx).Status = sentry.SpanStatusUnauthenticated
	c.updateAuthenticationCookie(ctx, ClearAuthentication)
	return c.wrapAndReturnError(ctx, err, http.StatusUnauthorized, "unauthorized")
}

func (c *Controller) badRequest(ctx echo.Context, msg string, args ...any) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.returnError(ctx, http.StatusBadRequest, msg, args...)
}

func (c *Controller) badRequestError(ctx echo.Context, err error, msg string, args ...any) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, msg, args...)
}

func (c *Controller) invalidJson(ctx echo.Context) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.returnError(ctx, http.StatusBadRequest, "invalid JSON body")
}

func (c *Controller) invalidJsonError(ctx echo.Context, err error) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid JSON body")
}

func (c *Controller) notFound(ctx echo.Context, msg string, args ...any) error {
	return c.returnError(ctx, http.StatusNotFound, msg, args...)
}
