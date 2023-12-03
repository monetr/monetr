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
	"sync"
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
	var wg sync.WaitGroup
	defer wg.Wait()
	return filepath.WalkDir(filePath, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			wg.Add(1)
			go func(filePath string) {
				defer wg.Done()
				findTestsInFile(token.NewFileSet(), filePath)
			}(filePath)
		}
		return nil
	})
}

func findTestsInFile(fset *token.FileSet, filename string) error {
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, f := range node.Decls {
		wg.Add(1)
		go func(f ast.Decl) {
			defer wg.Done()
			if fn, ok := f.(*ast.FuncDecl); ok {
				if fn.Name.IsExported() && strings.HasPrefix(fn.Name.Name, "Test") {
					if len(fn.Type.Params.List) != 1 {
						return
					}

					arg := fn.Type.Params.List[0]
					if t, ok := arg.Type.(*ast.StarExpr); ok {
						if s, ok := t.X.(*ast.SelectorExpr); ok {
							if pkg, ok := s.X.(*ast.Ident); ok {
								if pkg.Name != "testing" {
									return
								}
							}

							if s.Sel.Name != "T" {
								return
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

						// fmt.Printf("%s|%s|%s\n", filename, fn.Name.Name, strings.Trim(name.Value, `"`))
						fmt.Printf("%s|%s\n", fn.Name.Name, strings.Trim(name.Value, `"`))
						//fmt.Println(fn.Name.Name, name.Value)
						//fmt.Printf("^\\Q%s\\E$^\\Q/%s\\E$\n", fn.Name.Name, strings.ReplaceAll(strings.Trim(name.Value, `"`), " ", "_"))

					}
					if !foundSubTests {
						// ^\QTestCleanupJobsJob_Run\E$
						//fmt.Printf("^\\Q%s\\E$\n", fn.Name.Name)
						// fmt.Printf("%s|%s\n", filename, fn.Name.Name)
						fmt.Println(fn.Name.Name)
					}
				}
			}
		}(f)
	}
	wg.Wait()
	return nil
}
