package testutils

import (
	"fmt"
	"reflect"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	_ gomock.Matcher = &GenericMatcher[any]{}
)

type GenericMatcher[T any] struct {
	callback func(data T) bool
}

// Matches implements gomock.Matcher
func (g *GenericMatcher[T]) Matches(x any) bool {
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

type equalValuesMatcher struct {
	expected any
}

// Matches implements gomock.Matcher.
func (e *equalValuesMatcher) Matches(received any) bool {
	return assert.ObjectsAreEqualValues(
		dereference(e.expected),
		dereference(received),
	)
}

func dereference[T any](input T) any {
	reflection := reflect.ValueOf(input)
	switch reflection.Kind() {
	case reflect.Array, reflect.Ptr:
		if reflection.IsNil() {
			return nil
		}
	}

	if reflection.Kind() == reflect.Ptr {
		return dereference(reflection.Elem().Interface())
	}

	return input
}

// String implements gomock.Matcher.
func (e *equalValuesMatcher) String() string {
	value := dereference(e.expected)
	return fmt.Sprintf("is equal enough to %v (%T)", value, value)
}

// EqVal is a gomock matcher that will compare the values provided as values
// alone. It does this by dereferencing pointers if they are not nil until the
// base value is found and then using assert.ObjectsAreEqualValues to perform
// the comparison.
func EqVal(expected any) gomock.Matcher {
	return &equalValuesMatcher{
		expected: expected,
	}
}
