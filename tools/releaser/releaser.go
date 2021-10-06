package main

import (
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5"
	"path/filepath"
	"time"
)

var (
	repoDirectory = flag.String("directory", "./", "repository directory for release")
)

func main() {
	flag.Parse()

	path, err := filepath.Abs(*repoDirectory)
	if err != nil {
		panic(err)
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	since := time.Now().Add(-24 * time.Hour)

	iter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		Until: &since,
	})
	if err != nil {
		panic(err)
	}

	commit, err := iter.Next()
	if err != nil {
		panic(err)
	}

	fmt.Println(commit.Hash)
}
