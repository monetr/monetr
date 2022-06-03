package controller

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/pkg/errors"
)

// wrapPgError will wrap and return an error to the client. But will try to infer a status code from the error it is
// given. If it cannot infer a status code, an InternalServerError is used.
func (c *Controller) wrapPgError(ctx iris.Context, err error, msg string, args ...interface{}) {

	switch errors.Cause(err) {
	case pg.ErrNoRows:
		ctx.SetErr(errors.Errorf("%s: record does not exist", fmt.Sprintf(msg, args...)))
		ctx.StatusCode(http.StatusNotFound)
		ctx.StopExecution()

		crumbs.Error(c.getContext(ctx), fmt.Sprintf(msg, args...), c.configuration.APIDomainName, map[string]interface{}{
			"error": ctx.GetErr().Error(),
		})
	default:
		switch actualErr := errors.Cause(err).(type) {
		case pg.Error:
			cleanedErr, status := c.sanitizePgError(actualErr)
			c.wrapAndReturnError(ctx, cleanedErr, status, msg, args...)
		default:
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, msg, args...)
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

func (c *Controller) wrapAndReturnError(ctx iris.Context, err error, status int, msg string, args ...interface{}) {
	ctx.SetErr(errors.Wrapf(err, msg, args...))
	ctx.StatusCode(status)
	ctx.StopExecution()

	crumbs.Error(c.getContext(ctx), fmt.Sprintf(msg, args...), c.configuration.APIDomainName, map[string]interface{}{
		"error": ctx.GetErr().Error(),
	})
}

func (c *Controller) failure(ctx iris.Context, status int, error GenericAPIError) {
	ctx.SetErr(error)
	ctx.StatusCode(status)
	ctx.StopExecution()

	crumbs.Error(c.getContext(ctx), error.FriendlyMessage(), c.configuration.APIDomainName, map[string]interface{}{
		"error": error,
	})
}

func (c *Controller) returnError(ctx iris.Context, status int, msg string, args ...interface{}) {
	ctx.SetErr(errors.Errorf(msg, args...))
	ctx.StatusCode(status)
	ctx.StopExecution()

	crumbs.Error(c.getContext(ctx), fmt.Sprintf(msg, args...), c.configuration.APIDomainName, map[string]interface{}{
		"error": ctx.GetErr().Error(),
	})
}

func (c *Controller) badRequest(ctx iris.Context, msg string, args ...interface{}) {
	c.returnError(ctx, http.StatusBadRequest, msg, args...)
}

func (c *Controller) invalidJson(ctx iris.Context) {
	c.returnError(ctx, http.StatusBadRequest, "invalid JSON body")
}

func (c *Controller) notFound(ctx iris.Context, msg string, args ...interface{}) {
	c.returnError(ctx, http.StatusNotFound, msg, args...)
}
