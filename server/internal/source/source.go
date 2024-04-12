package source

import (
	"embed"
	"fmt"
	"runtime"
	"strings"
)

//go:embed embed/**
var sourceCode embed.FS

var basePath string

func init() {
	pc, _, _, ok := runtime.Caller(0)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		path, _ := details.FileLine(pc)
		basePath = strings.TrimSuffix(path, "/server/internal/source/source.go")
	}
}

func transformFileName(path string) string {
	newPath := strings.TrimPrefix(path, basePath)
	return fmt.Sprintf("/embed/%s.txt", newPath)
}
