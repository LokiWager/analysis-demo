package typechecker

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// TypeChecker is an interface that defines a method that the checker should implement.
type TypeChecker interface {
	// Check checks the value of the interface and returns an error if the value is invalid.
	Check(interface{}) error
}

var CheckerAnalyzer = &analysis.Analyzer{
	Name: "checker",
	Doc:  "Analyze @check annotations and verify code correctness",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.VAR {
				return true
			}

			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, name := range valueSpec.Names {
					if name.Obj == nil {
						continue
					}
					if name.Obj.Kind != ast.Var {
						continue
					}

					if decl.Doc != nil {
						for _, comment := range decl.Doc.List {
							// split // and trim spaces
							comment.Text = strings.TrimPrefix(comment.Text, "//")
							comment.Text = strings.TrimSpace(comment.Text)
							checkerName, params, err := ParseComment(comment.Text)
							if err == nil {
								specValue := valueSpec.Values[i]
								value, err := ExtractValue(specValue)

								if err == nil {
									err := RunChecker(value, checkerName, params)
									if err != nil {
										log.Printf("Checker %s failed for %s: %v", checkerName, name.Name, err)
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

func ExtractValue(expr ast.Expr) (interface{}, error) {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return parseBasicLit(v)
	case *ast.Ident:
		if v.Name == "nil" {
			return nil, nil
		}
		return v.Name, nil
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", v)
	}
}

func parseBasicLit(lit *ast.BasicLit) (interface{}, error) {
	switch lit.Kind {
	case token.STRING:
		return strconv.Unquote(lit.Value)
	case token.INT:
		return strconv.Atoi(lit.Value)
	case token.FLOAT:
		return strconv.ParseFloat(lit.Value, 64)
	case token.CHAR:
		return strconv.Unquote(lit.Value)
	default:
		return nil, fmt.Errorf("unsupported literal type: %s", lit.Kind)
	}
}
