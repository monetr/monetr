package testutils

import (
	"github.com/golang/mock/gomock"
)

var (
	_ gomock.Matcher = &GenericMatcher[interface{}]{}
)

type GenericMatcher[T any] struct {
	callback func(data T) bool
}

// Matches implements gomock.Matcher
func (g *GenericMatcher[T]) Matches(x interface{}) bool {
	switch t := x.(type) {
	case T:
		return g.callback(t)
	default:
		return false
	}
}

// String implements gomock.Matcher
func (g *GenericMatcher[T]) String() string {
	return "Input that matches provided predicate!"
}

func NewGenericMatcher[T any](callback func(data T) bool) gomock.Matcher {
	return &GenericMatcher[T]{
		callback: callback,
	}
}
