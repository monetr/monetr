package router

import (
	"context"
	iris_context "github.com/kataras/iris/v12/context"
	"github.com/monetrapp/rest-api/pkg/repository"
)

type Handler func(ctx Context) (interface{}, error)

type Context interface {
	context.Context
	Next()
	Params() *iris_context.RequestParams
	ReadJSON(destination interface{})
	MustReadJSON(destination interface{})
	JSON(response interface{})
	UserID() uint64
	AccountID() uint64
	LoginID() uint64
	UnauthenticatedRepo() repository.UnauthenticatedRepository
	Repository() repository.Repository
}

type Router interface {
	Use(handler Handler)
	OnAnyErrorCode(handler Handler)
	OnErrorCode(errorCode int, handler Handler)
	PartyFunc(relativePath string, partyHandler func(party Router))

	Get(relativePath string, handler Handler)
	Post(relativePath string, handler Handler)
	Put(relativePath string, handler Handler)
	Delete(relativePath string, handler Handler)
	Patch(relativePath string, handler Handler)
}
