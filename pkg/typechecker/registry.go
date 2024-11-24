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
	"fmt"
	"strings"
)

var CheckerRegistry = map[string]TypeCheckerFactory{}

type TypeCheckerFactory func(params []string) TypeChecker

func RegisterChecker(name string, factory TypeCheckerFactory) {
	CheckerRegistry[name] = factory
}

func ParseComment(comment string) (name string, params []string, err error) {
	if !strings.HasPrefix(comment, "@check:") {
		return "", nil, errors.New("not a valid checker comment")
	}

	parts := strings.Split(strings.TrimPrefix(comment, "@check:"), ":")
	name = parts[0]
	if len(parts) > 1 {
		params = strings.Split(parts[1], ",")
	}
	return name, params, nil
}

func RunChecker(value interface{}, name string, params []string) error {
	factory, exists := CheckerRegistry[name]
	if !exists {
		return fmt.Errorf("checker %s not found", name)
	}
	checker := factory(params)
	return checker.Check(value)
}
