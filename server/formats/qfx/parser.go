package qfx

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	headerRegex  = regexp.MustCompile(`^(?:^\w+:\w+$)+$`)
	payloadRegex = regexp.MustCompile(`^<OFX>([\W|\w|\s]+)</OFX>$`)
	dataRegex    = regexp.MustCompile(`(?P<tag><[\w|\S]+>)(?P<value>.+)?`)
)

type ItemType uint8

const (
	ArrayStartItemType ItemType = 0
	ArrayEndItemType   ItemType = 1
	FieldItemType      ItemType = 2
)

type Token interface {
	Token()
}

type Field struct {
	Name  string
	Value string
}

func (Field) Token() {}

type Array struct {
	Name  string
	Items []Token
}

func (Array) Token() {}

func Parse(qfxData string) Token {
	items := dataRegex.FindAllStringSubmatch(qfxData, -1)
	_, token := parseItem(0, items)
	return token
}

func parseItem(index int, items [][]string) (i int, result Token) {
	item := items[index]
	switch getItemType(item) {
	case ArrayStartItemType:
		return parseArray(index, items)
	case FieldItemType:
		return parseField(index, items)
	default:
		panic(fmt.Sprintf("syntax error at index [%d]", index))
	}
}

func getItemType(item []string) ItemType {
	value := item[2]
	name := item[1]
	if value == "" {
		isClosing := strings.HasPrefix(name, "</")
		if isClosing {
			return ArrayEndItemType
		}
		return ArrayStartItemType
	}

	return FieldItemType
}

func parseArray(index int, items [][]string) (i int, result Token) {
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
			i, tmp = parseItem(i, items)
			token.Items = append(token.Items, tmp)
		}
	}
	return i, token
}

func parseField(index int, items [][]string) (i int, result Token) {
	return index, &Field{
		Name:  cleanName(items[index][1]),
		Value: items[index][2],
	}
}

func cleanName(name string) string {
	return strings.Trim(name, "<>")
}
