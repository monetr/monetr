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
		first := bytes.TrimPrefix(parts[0], prefix)
		last := bytes.TrimSuffix(parts[1], suffix)
		result[string(first)] = string(last)
	}

	return result, nil
}
