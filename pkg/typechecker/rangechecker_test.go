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

package typechecker

import "testing"

func TestChecker_Range(t *testing.T) {
	RegisterChecker("Range", func(params []string) TypeChecker {
		minN := 10.0
		maxN := 100.0
		return NewRangeChecker(minN, maxN)
	})

	err := RunChecker(50, "Range", []string{"10", "100"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = RunChecker(200, "Range", []string{"10", "100"})
	if err == nil {
		t.Errorf("Expected error for out-of-range value, got none")
	}

	err = RunChecker(10, "Range", []string{"10", "100"})
	if err != nil {
		t.Errorf("Expected no error for lower boundary, got: %v", err)
	}

	err = RunChecker(100, "Range", []string{"10", "100"})
	if err != nil {
		t.Errorf("Expected no error for upper boundary, got: %v", err)
	}

}
