package icons

import (
	"bufio"
	"bytes"
	"io"
)

func parseSimpleIconsSlugMarkdown(data []byte) (map[string]string, error) {
	prefix := []byte("| `")
	middle := []byte("` | `")
	suffix := []byte("` |")
	buf := bufio.NewReader(bytes.NewBuffer(data))
	result := map[string]string{}
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}

		if !bytes.HasPrefix(line, prefix) {
			continue
		}

		parts := bytes.SplitN(line, middle, 2)
		first := bytes.ToLower(bytes.TrimPrefix(parts[0], prefix))
		firstNoSpaces := bytes.ReplaceAll(first, []byte(" "), []byte{})
		last := bytes.TrimSuffix(parts[1], suffix)
		result[string(first)] = string(last)

		if !bytes.Equal(first, firstNoSpaces) {
			result[string(firstNoSpaces)] = string(last)
		}
	}

	return result, nil
}
