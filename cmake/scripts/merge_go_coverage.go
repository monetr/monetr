package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"golang.org/x/tools/cover"
)

type blockKey struct {
	File      string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s output.txt input1.txt input2.txt ...\n", os.Args[0])
		os.Exit(1)
	}

	outputFile := os.Args[1]
	inputFiles := os.Args[2:]

	var mode string
	merged := make(map[blockKey]int)

	for _, path := range inputFiles {
		profiles, err := cover.ParseProfiles(path)
		if err != nil {
			log.Fatalf("failed to parse %s: %v", path, err)
		}

		for _, p := range profiles {
			if mode == "" {
				mode = p.Mode
			} else if mode != p.Mode {
				log.Fatalf("coverage mode mismatch: %s vs %s", mode, p.Mode)
			}

			for _, b := range p.Blocks {
				key := blockKey{
					File:      p.FileName,
					StartLine: b.StartLine,
					StartCol:  b.StartCol,
					EndLine:   b.EndLine,
					EndCol:    b.EndCol,
					NumStmt:   b.NumStmt,
				}

				if mode == "set" {
					// For set mode, treat as boolean OR
					if b.Count > 0 {
						merged[key] = 1
					}
				} else {
					// count / atomic → additive merge
					merged[key] += b.Count
				}
			}
		}
	}

	if err := writeMerged(outputFile, mode, merged); err != nil {
		log.Fatal(err)
	}
}

func writeMerged(path, mode string, merged map[blockKey]int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "mode: %s\n", mode)

	// Sort for deterministic output
	var keys []blockKey
	for k := range merged {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].File != keys[j].File {
			return keys[i].File < keys[j].File
		}
		if keys[i].StartLine != keys[j].StartLine {
			return keys[i].StartLine < keys[j].StartLine
		}
		if keys[i].StartCol != keys[j].StartCol {
			return keys[i].StartCol < keys[j].StartCol
		}
		if keys[i].EndLine != keys[j].EndLine {
			return keys[i].EndLine < keys[j].EndLine
		}
		return keys[i].EndCol < keys[j].EndCol
	})

	for _, k := range keys {
		fmt.Fprintf(
			f,
			"%s:%d.%d,%d.%d %d %d\n",
			k.File,
			k.StartLine,
			k.StartCol,
			k.EndLine,
			k.EndCol,
			k.NumStmt,
			merged[k],
		)
	}

	return nil
}
