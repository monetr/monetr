package ofx

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	headerRegex  = regexp.MustCompile(`^(?:^\w+:\w+$)+$`)
	payloadRegex = regexp.MustCompile(`^<OFX>([\W|\w|\s]+)</OFX>$`)
	dataRegex    = regexp.MustCompile(`(?P<tag><[/a-zA-Z0-9.]+>)(?P<value>[^<]+)?`)
)

type ItemType uint8

const (
	ArrayStartItemType ItemType = 0
	ArrayEndItemType   ItemType = 1
	FieldItemType      ItemType = 2
)

type Token interface {
	Token() string
	XML() string
}

type Field struct {
	Name  string
	Value string
}

func (f Field) Token() string {
	return f.Name
}

func (f Field) XML() string {
	return fmt.Sprintf("<%s>%s</%s>", f.Name, strings.TrimSpace(f.Value), f.Name)
}

type Array struct {
	Name  string
	Items []Token
}

func (a Array) Token() string {
	return a.Name
}

func (a Array) XML() string {
	pieces := make([]string, len(a.Items))
	for i := range a.Items {
		pieces[i] = a.Items[i].XML()
	}
	return fmt.Sprintf("<%s>%s</%s>", a.Name, strings.Join(pieces, ""), a.Name)
}

func Tokenize(ofxData string) (Token, error) {
	items := dataRegex.FindAllStringSubmatch(ofxData, -1)
	if len(items) == 0 {
		return nil, errors.New("OFX file provided is not valid")
	}
	_, token := tokenizeItem(0, items)
	return token, nil
}

func tokenizeItem(index int, items [][]string) (i int, result Token) {
	item := items[index]
	switch getItemType(item) {
	case ArrayStartItemType:
		return tokenizeArray(index, items)
	case FieldItemType:
		return tokenizeField(index, items)
	default:
		panic(fmt.Sprintf("syntax error at index [%d]", index))
	}
}

func getItemType(item []string) ItemType {
	value := strings.TrimSpace(item[2])
	name := strings.TrimSpace(item[1])
	if value == "" {
		isClosing := strings.HasPrefix(name, "</")
		if isClosing {
			return ArrayEndItemType
		}
		return ArrayStartItemType
	}

	return FieldItemType
}

func tokenizeArray(index int, items [][]string) (i int, result Token) {
	var token *Array
	for i = index; i < len(items); i++ {
		item := items[i]
		name := item[1]
		if token == nil {
			token = &Array{
				Name:  cleanName(name),
				Items: make([]Token, 0),
			}
			continue
		}

		switch getItemType(item) {
		case ArrayEndItemType:
			return i, token
		default:
			var tmp Token
			i, tmp = tokenizeItem(i, items)
			token.Items = append(token.Items, tmp)
		}
	}
	return i, token
}

func tokenizeField(index int, items [][]string) (i int, result Token) {
	return index, &Field{
		Name:  cleanName(items[index][1]),
		Value: items[index][2],
	}
}

func cleanName(name string) string {
	return strings.Trim(strings.TrimSpace(name), "<>")
}
