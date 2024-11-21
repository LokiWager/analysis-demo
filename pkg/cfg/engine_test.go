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
	"testing"

	testAssert "github.com/stretchr/testify/assert"
)

func TestEngine_ForIfControl(t *testing.T) {
	t.Run("Test Engine for if control", func(t *testing.T) {
		assert := testAssert.New(t)
		engine := NewEngine("../../tests/control_if", "example.go", nil)
		if engine == nil {
			t.Errorf("new engine failed")
			return
		}
		err := engine.CreateProgram()
		if err != nil {
			t.Errorf("create program failed: %v", err)
		}

		assert.Len(engine.result, 2)
		assert.Equal("x * 4:int", engine.result[0].Instr)
		assert.Equal("Even", engine.result[0].Parity)
		assert.Equal("x + 5:int", engine.result[1].Instr)
		assert.Equal("⊤", engine.result[1].Parity)
	})
}

func TestEngine_ForForControl(t *testing.T) {
	t.Run("Test Engine for for control", func(t *testing.T) {
		assert := testAssert.New(t)
		engine := NewEngine("../../tests/control_for", "example.go", nil)
		if engine == nil {
			t.Errorf("new engine failed")
			return
		}
		err := engine.CreateProgram()
		if err != nil {
			t.Errorf("create program failed: %v", err)
		}

		assert.Len(engine.result, 2)
		assert.Equal("x * 2:int", engine.result[0].Instr)
		assert.Equal("Even", engine.result[0].Parity)
		assert.Equal("i + 1:int", engine.result[1].Instr)
		assert.Equal("⊤", engine.result[1].Parity)
	})
}

func TestEngine_ForFuncCall(t *testing.T) {
	t.Run("Test Engine for function call", func(t *testing.T) {
		assert := testAssert.New(t)
		engine := NewEngine("../../tests/func_call", "example.go", nil)
		if engine == nil {
			t.Errorf("new engine failed")
			return
		}
		err := engine.CreateProgram()
		if err != nil {
			t.Errorf("create program failed: %v", err)
		}

		assert.Len(engine.result, 5)
		assert.Equal("x * 4:int", engine.result[0].Instr)
		assert.Equal("Even", engine.result[0].Parity)
		assert.Equal("y + 1:int", engine.result[1].Instr)
		assert.Equal("⊤", engine.result[1].Parity)
		assert.Equal("i + 1:int", engine.result[2].Instr)
		assert.Equal("⊤", engine.result[2].Parity)
		assert.Equal("y * 2:int", engine.result[3].Instr)
		assert.Equal("Even", engine.result[3].Parity)
		assert.Equal("x + y * 2:int", engine.result[4].Instr)
		assert.Equal("⊤", engine.result[1].Parity)
	})
}
