package ofx

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
)

var (
	dataRegex = regexp.MustCompile(`(?P<tag><[/a-zA-Z0-9.]+>)(?P<value>[^<]+)?`)
)

const (
	// maxDepth is how deep arrays can be nested in an OFX file before we give
	// up. Real files are only like 10 levels deep so this is plenty of room.
	maxDepth = 64
)

type ItemType uint8

const (
	ArrayStartItemType ItemType = 0
	ArrayEndItemType   ItemType = 1
	FieldItemType      ItemType = 2
)

type Token interface {
	Token() []byte
	writeXML(buf *bytes.Buffer)
}

type Field struct {
	Name  []byte
	Value []byte
}

func (f Field) Token() []byte {
	return f.Name
}

func (f Field) writeXML(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "<%s>%s</%s>", f.Name, string(bytes.TrimSpace(f.Value)), f.Name)
}

type Array struct {
	Name  []byte
	Items []Token
}

func (a Array) Token() []byte {
	return a.Name
}

// writeXML is recursive, but that's fine because Tokenize won't let the arrays
// get nested deeper than maxDepth. Everything gets written to the same buffer
// so we aren't copying the children over and over again at every level.
func (a Array) writeXML(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "<%s>", a.Name)
	for i := range a.Items {
		a.Items[i].writeXML(buf)
	}
	fmt.Fprintf(buf, "</%s>", a.Name)
}

// Tokenize will walk the OFX data one tag at a time and build the tree of
// arrays and fields. We keep our own stack of the open arrays instead of
// recursing, otherwise a file that was just <A> over and over again would
// recurse once per tag and take forever (or blow the stack). We also bail if
// the stack gets deeper than maxDepth.
func Tokenize(ctx context.Context, ofxData []byte) (Token, error) {
	var root Token
	stack := make([]*Array, 0, 16)
	for index, offset := 0, 0; ; index++ {
		// Once the root is closed we are done, anything after it gets ignored.
		if root != nil && len(stack) == 0 {
			break
		}

		// Don't check the context on every single tag, it's not free.
		if index%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, errors.Wrap(err, "failed to tokenize OFX data")
			}
		}

		match := dataRegex.FindSubmatchIndex(ofxData[offset:])
		if match == nil {
			break
		}
		item := getItem(ofxData[offset:], match)
		offset += match[1]

		switch getItemType(item) {
		case ArrayStartItemType:
			if len(stack) >= maxDepth {
				return nil, errors.Errorf("OFX data is nested too deep, more than [%d] levels at index [%d]", maxDepth, index)
			}
			array := &Array{
				Name:  cleanName(item[1]),
				Items: make([]Token, 0),
			}
			if len(stack) == 0 {
				root = array
			} else {
				parent := stack[len(stack)-1]
				parent.Items = append(parent.Items, array)
			}
			stack = append(stack, array)
		case FieldItemType:
			field := &Field{
				Name:  cleanName(item[1]),
				Value: item[2],
			}
			if len(stack) == 0 {
				root = field
			} else {
				parent := stack[len(stack)-1]
				parent.Items = append(parent.Items, field)
			}
		case ArrayEndItemType:
			// Closing tag with nothing open, the file starts with a closing tag.
			if len(stack) == 0 {
				return nil, errors.Errorf("syntax error at index [%d]", index)
			}
			stack = stack[:len(stack)-1]
		}
	}

	if root == nil {
		return nil, errors.New("OFX file provided is not valid")
	}

	// If there are still arrays open at the end of the file that's fine, we just
	// treat them as closed.
	return root, nil
}

// getItem takes the indexes from FindSubmatchIndex and turns them into the same
// shape that FindAllSubmatch would have given us. The value group is optional
// so it might be -1 if it didn't match anything.
func getItem(data []byte, match []int) [][]byte {
	item := make([][]byte, 3)
	item[0] = data[match[0]:match[1]]
	item[1] = data[match[2]:match[3]]
	if match[4] >= 0 {
		item[2] = data[match[4]:match[5]]
	}
	return item
}

func getItemType(item [][]byte) ItemType {
	value := bytes.TrimSpace(item[2])
	name := bytes.TrimSpace(item[1])
	if len(value) == 0 {
		isClosing := bytes.HasPrefix(name, []byte("</"))
		if isClosing {
			return ArrayEndItemType
		}
		return ArrayStartItemType
	}

	return FieldItemType
}

func cleanName(name []byte) []byte {
	return bytes.Trim(bytes.TrimSpace(name), "<>")
}
