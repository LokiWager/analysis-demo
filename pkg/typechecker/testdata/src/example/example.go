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

package example

import "fmt"

func main() {
	// @check:NotNullable
	var a = "Hello"

	// @check:Range:10,100
	var b = 50

	// @check:Range:10,100
	var c = 200

	// @check:MatchPattern:^[a-zA-Z0-9]+$
	var d = "HelloWorld"

	fmt.Println(a, b, c, d)
}
