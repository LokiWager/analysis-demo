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

package engine_test

import (
	"testing"

	testAssert "github.com/stretchr/testify/assert"

	"github.com/LokiWager/analysis-demo/pkg/engine"
)

func TestEngine_CheckIdentifiers(t *testing.T) {
	t.Run("CheckIdentifiers for vars", func(t *testing.T) {
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
		e := engine.NewEngine("", src)

		targetIdentifiers, ok := e.CheckIdentifiers()
		assert.True(ok)
		assert.Len(targetIdentifiers, 1)
		assert.Equal("idEqual13xxxx", targetIdentifiers[0].Name)
	})

	t.Run("CheckIdentifiers for functions", func(t *testing.T) {
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
		e := engine.NewEngine("", src)

		targetIdentifiers, ok := e.CheckIdentifiers()
		assert.True(ok)
		assert.Len(targetIdentifiers, 1)
		assert.Equal("idEqual13xxxx", targetIdentifiers[0].Name)
	})

	t.Run("CheckIdentifiers for structs", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

type idEqual13xxxx struct {
	Name string
}
`
		e := engine.NewEngine("", src)

		targetIdentifiers, ok := e.CheckIdentifiers()
		assert.True(ok)
		assert.Len(targetIdentifiers, 1)
	})

	t.Run("CheckIdentifiers for multiple identifiers", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func idEqual13xxxx() {
	anotherIdEqua := "abcdefghijklm"
	return
}
`
		e := engine.NewEngine("", src)

		targetIdentifiers, ok := e.CheckIdentifiers()
		assert.True(ok)
		assert.Len(targetIdentifiers, 2)
		assert.Equal("idEqual13xxxx", targetIdentifiers[0].Name)
		assert.Equal("anotherIdEqua", targetIdentifiers[1].Name)
	})

	t.Run("CheckIdentifiers for no identifiers", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	return
}
`
		e := engine.NewEngine("", src)

		targetIdentifiers, ok := e.CheckIdentifiers()
		assert.False(ok)
		assert.Len(targetIdentifiers, 0)
	})
}

func TestEngine_CheckControlFlow(t *testing.T) {
	t.Run("CheckControlFlow with if statement", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	if true {
		if false {
			if true {
				if false {
					return
				}
			}
		}
	}
}
`
		e := engine.NewEngine("", src)

		hasExceeded := e.CheckControlFlow()
		assert.True(hasExceeded)
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
					return
				}
			}
		}
	}
}
`
		e := engine.NewEngine("", src)

		hasExceeded := e.CheckControlFlow()
		assert.True(hasExceeded)
	})

	t.Run("CheckControlFlow with no control flow", func(t *testing.T) {
		assert := testAssert.New(t)
		src := `
package main

func main() {
	return
}
`
		e := engine.NewEngine("", src)

		hasExceeded := e.CheckControlFlow()
		assert.False(hasExceeded)
	})
}
