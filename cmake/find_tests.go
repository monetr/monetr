package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// testMap is a map of directories
type testMap map[string][][]string

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run find_tests.go <path>")
		return
	}

	path := os.Args[1]
	tests, err := findTestsInPath(path)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}

	output, err := json.Marshal(tests)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(output))
}

func findTestsInPath(filePath string) (testMap, error) {
	tests := testMap{}

	return tests, filepath.WalkDir(filePath, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			directory := path.Dir(filePath)
			testsFromFile, err := findTestsInFile(token.NewFileSet(), filePath)
			if err != nil {
				return err
			}

			dirTests := tests[directory]
			dirTests = append(dirTests, testsFromFile...)
			tests[directory] = dirTests
		}
		return nil
	})
}

func findTestsInFile(fset *token.FileSet, filename string) ([][]string, error) {
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	tests := make([][]string, 0, len(node.Decls))

	for _, f := range node.Decls {
		if fn, ok := f.(*ast.FuncDecl); ok {
			if fn.Name.IsExported() && strings.HasPrefix(fn.Name.Name, "Test") {
				if len(fn.Type.Params.List) != 1 {
					continue
				}

				arg := fn.Type.Params.List[0]
				if t, ok := arg.Type.(*ast.StarExpr); ok {
					if s, ok := t.X.(*ast.SelectorExpr); ok {
						if pkg, ok := s.X.(*ast.Ident); ok {
							if pkg.Name != "testing" {
								continue
							}
						}

						if s.Sel.Name != "T" {
							continue
						}
					}
				}

				foundSubTests := false
				for _, item := range fn.Body.List {
					stmt, ok := item.(*ast.ExprStmt)
					if !ok {
						continue
					}

					call, ok := stmt.X.(*ast.CallExpr)
					if !ok {
						continue
					}

					selector, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						continue
					}

					x, ok := selector.X.(*ast.Ident)
					if !ok {
						continue
					}
					if x.Name != "t" {
						continue
					}
					if selector.Sel.Name != "Run" {
						continue
					}

					name, ok := call.Args[0].(*ast.BasicLit)
					if !ok {
						continue
					}

					foundSubTests = true
					tests = append(tests, []string{
						fn.Name.Name,
						strings.Trim(name.Value, `"`),
					})
				}
				if !foundSubTests {
					tests = append(tests, []string{
						fn.Name.Name,
					})
				}
			}
		}
	}

	return tests, nil
}
