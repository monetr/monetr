package router

import (
	"context"
	"github.com/kataras/iris/v12"
	iris_context "github.com/kataras/iris/v12/context"
	"github.com/monetrapp/rest-api/pkg/repository"
	"time"
)

var (
	_ Context = &irisContext{}
	_ Router  = &irisRouter{}
)

type (
	irisRouter struct {
		app *iris.Application
	}

	irisContext struct {
		ctx     context.Context
		irisCtx iris.Context
	}
)

func (i *irisContext) Deadline() (deadline time.Time, ok bool) {
	return i.ctx.Deadline()
}

func (i *irisContext) Done() <-chan struct{} {
	return i.ctx.Done()
}

func (i *irisContext) Err() error {
	return i.ctx.Err()
}

func (i *irisContext) Value(key interface{}) interface{} {
	return i.ctx.Value(key)
}

func (i *irisContext) Next() {
	i.irisCtx.Next()
}

func (i *irisContext) Params() *iris_context.RequestParams {
	panic("implement me")
}

func (i *irisContext) ReadJSON(destination interface{}) {
	panic("implement me")
}

func (i *irisContext) MustReadJSON(destination interface{}) {
	panic("implement me")
}

func (i *irisContext) JSON(response interface{}) {
	panic("implement me")
}

func (i *irisContext) UserID() uint64 {
	panic("implement me")
}

func (i *irisContext) AccountID() uint64 {
	panic("implement me")
}

func (i *irisContext) LoginID() uint64 {
	panic("implement me")
}

func (i *irisContext) UnauthenticatedRepo() repository.UnauthenticatedRepository {
	panic("implement me")
}

func (i *irisContext) Repository() repository.Repository {
	panic("implement me")
}

func (i *irisRouter) Use(handler Handler) {
	i.app.Use(func(irisCtx iris.Context) {
		_, err := handler(i.contextFromIris(irisCtx))
		if err != nil {
			irisCtx.SetErr(err)
			return
		}
	})
}

func (i *irisRouter) OnAnyErrorCode(handler Handler) {
	panic("implement me")
}

func (i *irisRouter) OnErrorCode(errorCode int, handler Handler) {
	panic("implement me")
}

func (i *irisRouter) PartyFunc(relativePath string, partyHandler func(party Router)) {
	panic("implement me")
}

func (i *irisRouter) Get(relativePath string, handler Handler) {
	panic("implement me")
}

func (i *irisRouter) Post(relativePath string, handler Handler) {
	panic("implement me")
}

func (i *irisRouter) Put(relativePath string, handler Handler) {
	panic("implement me")
}

func (i *irisRouter) Delete(relativePath string, handler Handler) {
	panic("implement me")
}

func (i *irisRouter) Patch(relativePath string, handler Handler) {
	panic("implement me")
}

func NewRouter(app *iris.Application) Router {
	return &irisRouter{
		app: app,
	}
}

func (i *irisRouter) contextFromIris(irisCtx iris.Context) Context {
	betterContext, ok := irisCtx.Values().Get("_irisContext_").(*irisContext)
	if !ok {
		betterContext = &irisContext{
			ctx:     irisCtx.Request().Context(),
			irisCtx: irisCtx,
		}
		irisCtx.Values().Set("_irisContext_", betterContext)
	}

	return betterContext
}
