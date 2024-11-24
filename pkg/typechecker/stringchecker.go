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

import (
	"errors"
	"log"
	"regexp"
)

type StringPatternChecker struct {
	Pattern *regexp.Regexp
}

func NewStringPatternChecker(pattern string) (*StringPatternChecker, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &StringPatternChecker{Pattern: re}, nil
}

func (s *StringPatternChecker) Check(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("value is not a string")
	}
	if !s.Pattern.MatchString(str) {
		return errors.New("string does not match the pattern")
	}
	return nil
}

func MatchPattern(value interface{}, pattern string) interface{} {
	checker, err := NewStringPatternChecker(pattern)
	if err != nil {
		log.Printf("Failed to create pattern checker: %v", err)
		return nil
	}
	err = checker.Check(value)
	if err != nil {
		log.Printf("Pattern match failed: %v", err)
		return nil
	}
	return value
}

func init() {
	RegisterChecker("MatchPattern", func(params []string) TypeChecker {
		if len(params) != 1 {
			log.Fatalf("MatchPattern checker requires 1 parameter: pattern")
		}
		pattern := params[0]
		checker, err := NewStringPatternChecker(pattern)
		if err != nil {
			log.Fatalf("Invalid pattern: %v", err)
		}
		return checker
	})
}
