/*
 * Copyright (c) 2024, LokiWager
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cfg

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"log"

	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type (
	// Engine is the analysis engine
	Engine struct {
		// file set of the source code
		fileSet *token.FileSet

		// file of the source code
		file *ast.File

		// config of the source code
		conf *types.Config

		// prog of the source code
		prog *ssa.Program

		// pkgPath of the source code
		pkgPath string

		// result of the analysis
		result []analysisResult
	}

	analysisResult struct {
		Instr  string
		Parity string
	}
)

// NewEngine creates a new Engine instance
// path is the path of the source code, if not exists, pass an empty string
// src is the source code, if not exists, pass nil
// path and src must not be nil at the same time
func NewEngine(path string, fileName string, src any) *Engine {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, fmt.Sprintf("%s/%s", path, fileName), src, parser.AllErrors)
	if err != nil {
		logrus.Errorf("parse file %s failed: %v", path, err)
		panic(err)
	}

	return &Engine{
		fileSet: fileSet,
		file:    file,
		pkgPath: path,
	}
}

// GetPackage returns the package name of the source code
func (e *Engine) GetPackage() string {
	return e.file.Name.Name
}

// CreateProgram creates the program of the source code
func (e *Engine) CreateProgram() error {
	pkg := types.NewPackage(e.pkgPath, e.GetPackage())
	conf := types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	e.conf = &conf
	pkg, err := e.conf.Check(e.pkgPath, e.fileSet, []*ast.File{e.file}, info)
	if err != nil {
		logrus.Errorf("check file %s failed: %v", e.pkgPath, err)
		return err
	}
	e.pkgPath = pkg.Path()

	cfg := packages.Config{
		Mode: packages.LoadAllSyntax,
		Fset: e.fileSet,
	}
	initial, err := packages.Load(&cfg, e.pkgPath)
	if err != nil {
		log.Fatalf("load packages failed: %v", err)
		return err
	}
	prog, _ := ssautil.AllPackages(initial, ssa.SanityCheckFunctions)
	prog.Build()
	e.prog = prog

	for _, progPackage := range prog.AllPackages() {
		if progPackage.Pkg.Name() != pkg.Name() {
			continue
		}
		for _, member := range progPackage.Members {
			if fn, ok := member.(*ssa.Function); ok {
				if fn.Name() == "example" {
					e.analyzeEvenOdd(fn)
				}
			}
		}
	}

	return nil
}

// analyzeEvenOdd performs the even/odd analysis on a given function.
func (e *Engine) analyzeEvenOdd(fn *ssa.Function) string {
	var returnParity string = "⊤"             // Default return value is unknown (⊤)
	visited := make(map[*ssa.BasicBlock]bool) // Track visited blocks for fixed-point iteration

	for _, block := range fn.Blocks {
		e.analyzeBlock(block, visited)
	}
	return returnParity
}

// analyzeBlock handles analysis for each basic block, with branching and looping support.
func (e *Engine) analyzeBlock(block *ssa.BasicBlock, visited map[*ssa.BasicBlock]bool) {
	if visited[block] {
		return // Avoid re-analyzing blocks in loops (fixed-point reached).
	}
	visited[block] = true

	for _, instr := range block.Instrs {
		switch v := instr.(type) {
		case *ssa.BinOp:
			if v.Op != token.ADD && v.Op != token.MUL {
				continue // Skip non-addition and non-multiplication operations.
			}
			// Handle binary operations (e.g., addition and multiplication).
			left := e.analyzeValue(v.X)
			right := e.analyzeValue(v.Y)
			result := e.evenOddOperation(v.Op, left, right)
			e.result = append(e.result,
				analysisResult{
					Instr:  fmt.Sprintf("%s %s %s", e.getValueName(v.X), v.Op.String(), e.getValueName(v.Y)),
					Parity: result,
				},
			)

		case *ssa.If:
			// Analyze both branches in the condition.
			e.analyzeBlock(v.Block().Succs[0], visited)
			// Analyze false branch.
			e.analyzeBlock(v.Block().Succs[1], visited)

		case *ssa.Jump:
			// Continue to the next block.
			e.analyzeBlock(v.Block(), visited)

		case *ssa.Call:
			// Handle function calls.
			e.handleCall(v)
		}
	}
}

func (e *Engine) handleCall(call *ssa.Call) string {
	// Get the function being called.
	callee := call.Common().StaticCallee()

	// If it's a user-defined function (not a built-in or external package function), analyze it.
	if callee != nil {
		fmt.Printf("Analyzing called function: %s\n", callee.Name())
		return e.analyzeEvenOdd(callee) // Recursively analyze the called function.
	}

	// Handle built-in functions or external functions (e.g., fmt.Println).
	// For now, we assume unknown parity (⊤) for such cases.
	return "⊤"
}

// analyzeValue returns the parity (even, odd, unknown) for a given SSA value.
func (e *Engine) analyzeValue(val ssa.Value) string {
	switch v := val.(type) {
	case *ssa.Const:
		// Check if the constant is even or odd.
		if v.Value != nil && v.Value.Kind() == constant.Int {
			if v.Int64()%2 == 0 {
				return "Even"
			}
			return "Odd"
		}
	}
	// Default to unknown (⊤).
	return "⊤"
}

// evenOddOperation performs the even/odd analysis for binary operations.
func (e *Engine) evenOddOperation(op token.Token, left, right string) string {
	switch op {
	case token.ADD:
		if left == "Even" && right == "Even" {
			return "Even"
		}
		if left == "Odd" && right == "Odd" {
			return "Even"
		}
		if left == "Even" && right == "Odd" {
			return "Odd"
		}
		if left == "Odd" && right == "Even" {
			return "Odd"
		}
		return "⊤"
	case token.MUL:
		if left == "Even" || right == "Even" {
			return "Even"
		}
		if left == "Odd" && right == "Odd" {
			return "Odd"
		}
		return "⊤"
	default:
		return "⊤"
	}
}

func (e *Engine) getValueName(value ssa.Value) string {
	switch v := value.(type) {
	case *ssa.Phi:
		return v.Comment
	case *ssa.Const:
		return v.Name()
	case *ssa.BinOp:
		return fmt.Sprintf("%s %s %s", e.getValueName(v.X), v.Op.String(), e.getValueName(v.Y))
	}
	return value.Name()
}
