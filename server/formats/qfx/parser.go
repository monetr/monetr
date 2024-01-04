package qfx

import "github.com/pkg/errors"

type QFXData interface {
	Parse(token Token) error
}

func mustArray(current any, token Token) (*Array, error) {
	array, ok := token.(*Array)
	if !ok {
		return nil, errors.Errorf("%T token must be an array", current)
	}

	return array, nil
}

func Parse(token Token) (*OFX, error) {

}

func parseToken(input Token) interface{} {
	switch input.Token() {
	case "OFX":

	}
}
