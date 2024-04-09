package source

import "embed"

//go:embed embed/**
var sourceCode embed.FS
