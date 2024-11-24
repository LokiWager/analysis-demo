package checker

import (
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/analysis"

	"github.com/LokiWager/analysis-demo/pkg/typechecker"
)

var CheckerAnalyzer = &analysis.Analyzer{
	Name: "checker",
	Doc:  "Analyze @check annotations and verify code correctness",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.VAR {
				for _, spec := range decl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							if name.Obj != nil && name.Obj.Kind == ast.Var {
								if decl.Doc != nil {
									for _, comment := range decl.Doc.List {
										checkerName, params, err := typechecker.ParseComment(comment.Text)
										if err == nil {
											value := pass.TypesInfo.Types[name].Value
											if value != nil {
												err := typechecker.RunChecker(value, checkerName, params)
												if err != nil {
													log.Printf("Checker %s failed for %s: %v", checkerName, name.Name, err)
												}
											}
										}
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
