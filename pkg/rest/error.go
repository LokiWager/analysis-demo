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

package rest

import (
	"encoding/json"
	"fmt"

	echo "github.com/labstack/echo/v4"
)

type (
	// Err is the error format for all apis.
	Err struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// NewErr creates an error.
func NewErr(code int, message string) *Err {
	return &Err{
		Code:    code,
		Message: message,
	}
}

func (e *Err) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func handleErr(ctx echo.Context, err *Err) {
	buff, err1 := json.Marshal(err)
	if err1 != nil {
		panic(err1)
	}
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(err.Code)
	_, _ = ctx.Response().Write(buff)
}
