package util

import "reflect"

type MergeTransformer struct {
}

func (MergeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	switch typ.Kind() {
	case reflect.String:
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				srcValue := src.Interface().(string)
				dst.SetString(srcValue)
			}
			return nil
		}
	default:
		return nil
	}
}
