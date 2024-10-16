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

package ast

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/sirupsen/logrus"
)

type (
	// Engine is the analysis engine
	Engine struct {
		// file set of the source code
		fileSet *token.FileSet

		// file of the source code
		file *ast.File
	}
)

// NewEngine creates a new Engine instance
// path is the path of the source code, if not exists, pass an empty string
// src is the source code, if not exists, pass nil
// path and src must not be nil at the same time
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

// CheckIdentifiers checks if the identifiers' length is equal to 13
// returns true if all identifiers' length is not equal to 13, otherwise false
func (e *Engine) CheckIdentifiers() bool {
	// check if identifier's length is equal to 13
	noExists := false
	ast.Inspect(e.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if len(x.Name) == 13 {
				noExists = true
				return false
			}
		}

		return true
	})

	return !noExists
}

// CheckControlFlow checks if the control flow (if, for, switch, select) is nested more than 4 times
// returns true if control flow is not nested more than 4 times, otherwise false
func (e *Engine) CheckControlFlow() bool {
	// check if control flow (if, for, switch, select) is nested 4 times
	maxDepth := e.checkNestingLevel(e.file, 0, 4)
	return maxDepth <= 4
}

func (e *Engine) checkNestingLevel(parent ast.Node, depth, maxLevel int) int {
	if depth > maxLevel {
		return depth
	}

	maxDepth := depth

	ast.Inspect(parent, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.IfStmt:
			currentDepth := e.checkNestingLevel(x.Body, depth+1, maxLevel)
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}

			if x.Else != nil {
				currentDepth := e.checkNestingLevel(x.Else, depth+1, maxLevel)
				if currentDepth > maxDepth {
					maxDepth = currentDepth
				}
			}
		case *ast.ForStmt:
			currentDepth := e.checkNestingLevel(x.Body, depth+1, maxLevel)
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		case *ast.SwitchStmt:
			currentDepth := e.checkNestingLevel(x.Body, depth+1, maxLevel)
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		case *ast.SelectStmt:
			currentDepth := e.checkNestingLevel(x.Body, depth+1, maxLevel)
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		}
		return true
	})

	return maxDepth
}
