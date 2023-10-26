package mock_stripe

import (
	"net/url"
	"strings"
)

type StripeForm map[string]interface{}

func ParseStripeForm(input url.Values) (StripeForm, error) {
	data := map[string]interface{}{}

	for key, values := range input {
		var value interface{}
		switch len(values) {
		case 0:
			value = nil
		case 1:
			value = values[0]
		default:
			value = values
		}

		nextBracketIndex := strings.IndexRune(key, '[')
		if nextBracketIndex == -1 {
			data[key] = value
			continue
		}

		path := make([]string, 0)
		path = append(path, key[:nextBracketIndex])
		for {
			key = key[nextBracketIndex+1:]
			nextBracketIndex = strings.IndexRune(key, ']')

			part := key[:nextBracketIndex]
			path = append(path, part)

			key = key[nextBracketIndex+1:]
			nextBracketIndex = strings.IndexRune(key, '[')
			if nextBracketIndex == -1 {
				break
			}
		}

		part := data
		for i, item := range path {
			if i == len(path)-1 {
				part[item] = value
				continue
			}

			if _, ok := part[item]; !ok {
				part[item] = map[string]interface{}{}
			}

			part = part[item].(map[string]interface{})
		}
	}

	return data, nil
}
