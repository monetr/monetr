package controller

import (
	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
)

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
