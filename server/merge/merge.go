// Package merge comment
package merge

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	jsonUnmarshallerType = reflect.TypeFor[json.Unmarshaler]()
	jsonNumberType       = reflect.TypeFor[json.Number]()
	timeType             = reflect.TypeFor[time.Time]()
)

type MergeOption uint8

const (
	ErrorOnUnknownField MergeOption = 1
	SkipZeroValues      MergeOption = 2
)

func Merge[T any](dst *T, src map[string]any, options ...MergeOption) error {
	dstType := reflect.ValueOf(dst)
	var mergedOptions MergeOption
	for _, option := range options {
		mergedOptions |= option
	}

	m := &mergeContext{
		dst:     dstType,
		src:     src,
		options: mergedOptions,
	}

	return m.merge()
}

type mergeContext struct {
	dst        reflect.Value
	src        map[string]any
	options    MergeOption
	fields     []reflect.Value
	fieldsFold map[string]int
}

func (m *mergeContext) buildFieldMap(dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Ptr:
		// Most likely this will be a pointer first, so we go down a level in order
		// to work with the actual underlying type. Which _should_ be a struct in
		// this case.
		return m.buildFieldMap(dst.Elem())
	case reflect.Struct:
		numField := dst.NumField()
		m.fields = make([]reflect.Value, numField)
		m.fieldsFold = make(map[string]int, numField)
		for i := range numField {
			field := dst.Field(i)
			name := dst.Type().Field(i).Name
			m.fields[i] = field
			m.fieldsFold[strings.ToLower(name)] = i
		}
	default:
		return errors.Errorf("cannot build field map for destination of type: %s", dst.Type())
	}
	return nil
}

func (m *mergeContext) merge() error {
	if m.dst.IsNil() {
		return errors.New("cannot merge into a nil destination")
	}

	// Before we do anything we need to build our field map so we have something
	// to work with.
	if err := m.buildFieldMap(m.dst); err != nil {
		return err
	}

	for key, value := range m.src {
		keyFold := strings.ToLower(key)
		dstFieldIndex, ok := m.fieldsFold[keyFold]
		if !ok && m.options&ErrorOnUnknownField > 0 {
			return errors.Errorf("cannot assign field '%s' to destination", key)
		}
		dstField := m.fields[dstFieldIndex]

		srcValue := reflect.ValueOf(value)

		if m.options&SkipZeroValues > 0 && srcValue.IsZero() {
			continue
		}

		if srcValue.Kind() == reflect.Invalid {
			continue
		}

		switch {
		case dstField.Type().Implements(jsonUnmarshallerType) && srcValue.Kind() == reflect.String:
			// Create a new instance of the type of the pointer for the destination
			// field. For example if the dstField is `*RuleSet` then this will create
			// a new `RuleSet` instance and call UnmarshalJSON on it.
			value := reflect.New(dstField.Type().Elem())
			u := value.Interface().(json.Unmarshaler)
			if err := u.UnmarshalJSON([]byte(srcValue.String())); err != nil {
				return errors.WithStack(err)
			}
			// If that all works out then we can assign the resulting value to the
			// destination field even if it is nil.
			dstField.Set(value)
		case isInteger(dstField.Type()) && srcValue.Type() == jsonNumberType:
			// If the destination is an integer field and the source is a json number
			// then we can work with the destination field directly like this. This
			// way we can be better about how we parse and handle numbers.
			jsonNumber := srcValue.Interface().(json.Number)
			value, err := jsonNumber.Int64()
			if err != nil {
				return errors.WithStack(err)
			}

			// If the destination is a pointer then we need to set the inner value
			// instead of the value of the pointer.
			if dstField.Kind() == reflect.Pointer {
				newValue := reflect.New(reflect.TypeOf(value))
				newValue.Elem().Set(reflect.ValueOf(value))
				dstField.Set(newValue)
			} else {
				// Otherwise just set the value directly.
				dstField.SetInt(value)
			}
		case dstField.Kind() == reflect.String && srcValue.Kind() == reflect.String:
			// This is a weird very specific condition. But basically instead of using
			// Set() we can use SetString instead which does not perform the same type
			// checks. So if we are using a models.ID type for example, this will work
			// better for that.
			dstField.SetString(srcValue.String())
		case dstField.Type() == timeType && srcValue.Kind() == reflect.String:
			// If the destination is a timestamp and the source field is a string then
			// we should just parse the string!
			timestamp, err := time.Parse(time.RFC3339Nano, srcValue.String())
			if err != nil {
				return errors.WithStack(err)
			}
			dstField.Set(reflect.ValueOf(timestamp))
		case dstField.Kind() == srcValue.Kind():
			// If the destination and source are the exact same then we can just
			// assign the value directly.
			// TODO If they are both pointers but are of different types this will
			// fail.
			dstField.Set(srcValue)
		case dstField.Kind() == reflect.Pointer && srcValue.Kind() != reflect.Pointer:
			// If the destination is a pointer, but a pointer to the same type as the
			// source then we can create a new pointer value and assign.
			if dstField.Type().Elem().Kind() != srcValue.Kind() {
				// If the types do not match then we cannot handle them. The caller
				// would need to implement the UnmarshalJSON method in order to handle
				// custom type matching.
				return errors.Errorf("cannot assign field '%s', source is %s and destination is %s", key, srcValue.Type(), dstField.Type())
			}
			value := reflect.New(srcValue.Type())
			value.Elem().Set(srcValue)
			dstField.Set(value)
		default:
			return errors.Errorf("cannot assign field '%s', source is %s and destination is %s", key, srcValue.Type(), dstField.Type())
		}
	}

	return nil
}

func isInteger(val reflect.Type) bool {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Pointer:
		return isInteger(val.Elem())
	default:
		return false
	}
}
