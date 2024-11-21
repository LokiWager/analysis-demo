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

package ast_test

import (
	"testing"

	testAssert "github.com/stretchr/testify/assert"

	"github.com/LokiWager/analysis-demo/pkg/ast"
)

// TestEngine_Exists_CheckIdentifiers tests the CheckIdentifiers method of the Engine struct with true negative cases
func TestEngine_Exists_CheckIdentifiers(t *testing.T) {
	t.Run("Exists CheckIdentifiers for vars", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

import (
	"fmt"
)

func main() {
	idEqual13xxxx := "abcdefghijklm"
	idNotEqual13 := "abcdefghijk"
	
	fmt.Println(idEqual13)
	fmt.Println(idNotEqual13)
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckIdentifiers()
		assert.False(ok)
	})

	t.Run("Exists CheckIdentifiers for functions", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func idEqual13xxxx() {
	return
}

func idNotEqual13() {
	return
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckIdentifiers()
		assert.False(ok)
	})

	t.Run("Exists CheckIdentifiers for structs", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

type idEqual13xxxx struct {
	Name string
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckIdentifiers()
		assert.False(ok)
	})

	t.Run("Exists CheckIdentifiers for multiple identifiers", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func idEqual13xxxx() {
	anotherIdEqua := "abcdefghijklm"
	return
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckIdentifiers()
		assert.False(ok)
	})
}

// TestEngine_NoExists_CheckIdentifiers tests the CheckIdentifiers method of the Engine struct with true positive cases
func TestEngine_NoExists_CheckIdentifiers(t *testing.T) {
	t.Run("CheckIdentifiers for no identifiers", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	return
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckIdentifiers()
		assert.True(ok)
	})

	t.Run("CheckIdentifiers with no match", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	idNot13 := "abcdefghij"
	idNOT13 := "abcdefghijklmnop"
	fmt.Println(idNot13)
	fmt.Println(idNOT13)
}`
		e := ast.NewEngine("", src)

		ok := e.CheckIdentifiers()
		assert.True(ok)
	})
}

// TestEngine_Exists_CheckControlFlow tests the CheckControlFlow method of the Engine struct with true negative cases
func TestEngine_Exists_CheckControlFlow(t *testing.T) {
	t.Run("CheckControlFlow with if statement", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	if true {
		if false {
			if true {
				if false {
					if true {
						return
					}
				}
			}
		}
	}
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckControlFlow()
		assert.False(ok)
	})

	t.Run("CheckControlFlow with for, if, switch, select statement", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	for i := 0; i < 10; i++ {
		if true {
			switch i {
			case 1:
				select {
				default:
					if false {
						return
					}
					return
				}
			}
		}
	}
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckControlFlow()
		assert.False(ok)
	})
}

// TestEngine_NoExists_CheckControlFlow tests the CheckControlFlow method of the Engine struct with true positive cases
func TestEngine_NoExists_CheckControlFlow(t *testing.T) {
	t.Run("CheckControlFlow with no excessive control flow", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main
func main() {
	if false {
		return
	}
}
`
		e := ast.NewEngine("", src)
		ok := e.CheckControlFlow()
		assert.True(ok)
	})

	t.Run("CheckControlFlow with no excessive control flow", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main
func main() {
	for i := 0; i < 10; i++ {
	}
}
`
		e := ast.NewEngine("", src)
		ok := e.CheckControlFlow()
		assert.True(ok)
	})
}

// TestEngine_NoExists_CheckControlFlow tests the CheckControlFlow method of the Engine struct with false negative cases
func TestEngine_NoExists_FalsePositive_CheckControlFlow(t *testing.T) {
	t.Run("CheckControlFlow with no control flow", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	if true {
		if false {
			return
		}
	}
	return
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckControlFlow()
		assert.True(ok)
	})
}

// TestEngine_NoExists_CheckControlFlow tests the CheckControlFlow method of the Engine struct with false positive cases
func TestEngine_NoExists_FalseNegative_CheckControlFlow(t *testing.T) {
	t.Run("CheckControlFlow with no control flow", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	if true {
		return
	} else {
		if false {
			if true {
				if false {
					if true {
						return
					}
				}
			}
		}
	}
	return
}
`
		e := ast.NewEngine("", src)

		ok := e.CheckControlFlow()
		assert.False(ok)
	})
}
