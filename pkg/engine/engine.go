/*
 * Copyright (c) 2017, MegaEase
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

package engine

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/sirupsen/logrus"
)

type (
	Engine struct {
		// file set
		fileSet *token.FileSet

		// file
		file *ast.File
	}

	TargetIdentifiers struct {
		Name string
		Pos  token.Pos
	}
)

func NewEngine(path string, src any) *Engine {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, src, parser.ParseComments)
	if err != nil {
		logrus.Errorf("parse file %s failed: %v", path, err)
		panic(err)
	}

	return &Engine{
		fileSet: fileSet,
		file:    file,
	}
}

func (e *Engine) CheckIdentifiers() ([]TargetIdentifiers, bool) {
	// check if identifier's length is equal to 13
	targetIdentifiers := make([]TargetIdentifiers, 0)
	ast.Inspect(e.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if len(x.Name) == 13 {
				targetIdentifiers = append(targetIdentifiers, TargetIdentifiers{
					Name: x.Name,
					Pos:  x.Pos(),
				})
			}
		}

		return true
	})

	return targetIdentifiers, len(targetIdentifiers) > 0
}

func (e *Engine) CheckControlFlow() bool {
	// check if control flow (if, for, switch, select) is nested 4 times
	hasExceeded := false
	ast.Inspect(e.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.SwitchStmt, *ast.SelectStmt:
			if e.checkNestingLevel(x, 1, 4) {
				hasExceeded = true
				return false
			}
		}
		return true
	})

	return hasExceeded
}

func (e *Engine) checkNestingLevel(node ast.Node, depth, maxLevel int) bool {
	if depth == maxLevel {
		return true
	}

	switch x := node.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.SwitchStmt, *ast.SelectStmt:
		hasExceeded := false
		ast.Inspect(x, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.IfStmt, *ast.ForStmt, *ast.SwitchStmt, *ast.SelectStmt:
				hasExceeded = e.checkNestingLevel(n, depth+1, maxLevel)
				if hasExceeded {
					return false
				}
			}
			return true
		})
		return hasExceeded
	}
	return false
}
