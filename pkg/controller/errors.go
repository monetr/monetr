package controller

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
	"net/http"
)

// wrapPgError will wrap and return an error to the client. But will try to infer a status code from the error it is
// given. If it cannot infer a status code, an InternalServerError is used.
func (c *Controller) wrapPgError(ctx *context.Context, err error, msg string, args ...interface{}) {
	switch errors.Cause(err) {
	case pg.ErrNoRows:
		ctx.SetErr(errors.Errorf("%s: record does not exist", fmt.Sprintf(msg, args...)))
		ctx.StatusCode(http.StatusNotFound)
		ctx.StopExecution()
	default:
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, msg, args...)
	}
}

func (c *Controller) wrapAndReturnError(ctx *context.Context, err error, status int, msg string, args ...interface{}) {
	ctx.SetErr(errors.Wrapf(err, msg, args...))
	ctx.StatusCode(status)
	ctx.StopExecution()
}

func (c *Controller) returnError(ctx *context.Context, status int, msg string, args ...interface{}) {
	ctx.SetErr(errors.Errorf(msg, args...))
	ctx.StatusCode(status)
	ctx.StopExecution()
}

func (c *Controller) badRequest(ctx *context.Context, msg string, args ...interface{}) {
	c.returnError(ctx, http.StatusBadRequest, msg, args...)
}
