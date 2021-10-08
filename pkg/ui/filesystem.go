package ui

import "net/http"

var (
	_ http.FileSystem = &FileSystem{}
)

type FileSystem struct {
	prefix string
	fs     http.FileSystem
}

func NewFileSystem(prefix string, fs http.FileSystem) http.FileSystem {
	return &FileSystem{
		prefix: prefix,
		fs:     fs,
	}
}

func (f FileSystem) Open(name string) (http.File, error) {
	return f.fs.Open(f.prefix + name)
}
