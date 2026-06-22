package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
)

// apiResponseError carries an HTTP status code and a response body that should
// be written to the client verbatim. echo v4 let us smuggle this through
// HTTPError.Message (which was an interface{}) and HTTPError.Internal, but echo
// v5 narrowed Message down to a plain string and dropped Internal entirely, so
// monetr has to carry the shaped body itself now. The error middleware in
// routes.go detects this type and renders body as-is via ctx.JSON.
type apiResponseError struct {
	code int
	body any
	err  error
}

// StatusCode lets the error middleware decide the log level for this error the
// same way it does for echo.HTTPError.
func (e *apiResponseError) StatusCode() int {
	return e.code
}

func (e *apiResponseError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return fmt.Sprintf("api error with status code %d", e.code)
}

func (e *apiResponseError) Unwrap() error {
	return e.err
}

// wrapPgError will wrap and return an error to the client. But will try to infer a status code from the error it is
// given. If it cannot infer a status code, an InternalServerError is used.
func (c *Controller) wrapPgError(ctx *echo.Context, err error, msg string, args ...any) error {
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
		).Wrap(err)
	default:
		switch actualErr := errors.Cause(err).(type) {
		case pg.Error:
			status, cleanedErr := c.sanitizePgError(actualErr)
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

func (*Controller) sanitizePgError(err pg.Error) (int, error) {
	switch err.Field(67) {
	case "23505": // Duplicate
		// TODO Return actual duplicate information in this error.
		return http.StatusBadRequest, errors.New("a similar object already exists")
	default:
		return http.StatusInternalServerError, err
	}
}

func (c *Controller) wrapAndReturnError(ctx *echo.Context, err error, status int, msg string, args ...any) error {
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
		return echo.NewHTTPError(status, fmt.Sprintf(msg, args...)).Wrap(wrapped)
	}
}

func (c *Controller) failure(ctx *echo.Context, status int, apiError GenericAPIError) error {
	crumbs.Error(
		c.getContext(ctx),
		apiError.FriendlyMessage(),
		ctx.Request().URL.Hostname(),
		map[string]any{
			"error": apiError,
		},
	)

	// The GenericAPIError is itself a json.Marshaler, so we hand it through as the
	// response body and the error middleware renders it as-is, the same way echo
	// v4 did off of HTTPError.Internal.
	return &apiResponseError{
		code: status,
		body: apiError,
		err:  apiError,
	}
}

func (c *Controller) returnError(ctx *echo.Context, status int, msg string, args ...any) error {
	err := errors.Errorf(msg, args...)

	crumbs.Error(
		c.getContext(ctx),
		fmt.Sprintf(msg, args...),
		ctx.Request().URL.Hostname(),
		map[string]any{
			"error": err.Error(),
		},
	)

	return echo.NewHTTPError(status, fmt.Sprintf(msg, args...)).Wrap(err)
}

func (c *Controller) unauthorized(ctx *echo.Context) error {
	c.getSpan(ctx).Status = sentry.SpanStatusUnauthenticated
	c.updateAuthenticationCookie(ctx, ClearAuthentication)
	return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
}

func (c *Controller) unauthorizedError(ctx *echo.Context, err error) error {
	c.getSpan(ctx).Status = sentry.SpanStatusUnauthenticated
	c.updateAuthenticationCookie(ctx, ClearAuthentication)
	return c.wrapAndReturnError(ctx, err, http.StatusUnauthorized, "unauthorized")
}

func (c *Controller) badRequest(ctx *echo.Context, msg string, args ...any) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.returnError(ctx, http.StatusBadRequest, msg, args...)
}

func (c *Controller) badRequestError(ctx *echo.Context, err error, msg string, args ...any) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, msg, args...)
}

func (c *Controller) invalidJson(ctx *echo.Context) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.returnError(ctx, http.StatusBadRequest, "invalid JSON body")
}

func (c *Controller) invalidJsonError(ctx *echo.Context, err error) error {
	requestSpan := c.getSpan(ctx)
	requestSpan.Status = sentry.SpanStatusInvalidArgument
	return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid JSON body")
}

func (c *Controller) notFound(ctx *echo.Context, msg string, args ...any) error {
	return c.returnError(ctx, http.StatusNotFound, msg, args...)
}

// jsonError returns an error whose body is written verbatim from the provided
// map. The error middleware in routes.go detects the apiResponseError type and
// emits its body as-is, so callers can shape the response body (e.g. validation
// responses with a "problems" tree) without writing the JSON inline. Returning
// the error lets the handler short-circuit cleanly: subsequent code in the
// handler does not run, which avoids the double-body bug seen when ctx.JSON is
// called from a helper that then returns nil.
func (*Controller) jsonError(_ *echo.Context, status int, body map[string]any) error {
	return &apiResponseError{
		code: status,
		body: body,
	}
}
