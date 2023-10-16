package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run find_tests.go <path>")
		return
	}

	path := os.Args[1]
	err := findTestsInPath(path)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func findTestsInPath(filePath string) error {
	fset := token.NewFileSet()

	return filepath.WalkDir(filePath, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			err := findTestsInFile(fset, filePath)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func findTestsInFile(fset *token.FileSet, filename string) error {
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

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
					// ^\QTestGetSpending\E$/^\Qinvalid_bank_account_ID\E$

					fmt.Printf("%s|%s\n", fn.Name.Name, strings.Trim(name.Value, `"`))
					//fmt.Println(fn.Name.Name, name.Value)
					//fmt.Printf("^\\Q%s\\E$^\\Q/%s\\E$\n", fn.Name.Name, strings.ReplaceAll(strings.Trim(name.Value, `"`), " ", "_"))

				}
				if !foundSubTests {
					// ^\QTestCleanupJobsJob_Run\E$
					//fmt.Printf("^\\Q%s\\E$\n", fn.Name.Name)
					fmt.Println(fn.Name.Name)
				}
			}
		}
	}
	return nil
}
