package source

import (
	"embed"
	"fmt"
	"path"
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

func transformFileName(filePath string) string {
	newPath := strings.TrimPrefix(filePath, basePath)
	return fmt.Sprintf("%s.txt", path.Join("embed", newPath))
}
