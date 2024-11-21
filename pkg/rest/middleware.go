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
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/utils/timetool"
)

func newLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			startTime := time.Now()
			err := next(ctx)
			if err != nil {
				ctx.Error(err)
			}
			processTime := time.Now().Sub(startTime)

			method := ctx.Request().Method
			remoteAddr := ctx.Request().RemoteAddr
			path := ctx.Request().URL.Path
			code := ctx.Response().Status
			bodyBytesReceived := ctx.Request().ContentLength
			bodyBytesSent := ctx.Response().Size

			entry := fmt.Sprintf("%s %s %s %v rx:%dB tx:%dB start:%v process:%v",
				remoteAddr, method, path, code,
				bodyBytesReceived, bodyBytesSent,
				startTime.Format(timetool.RFC3339Milli), processTime)

			// NOTICE: Maybe separate it off the standard log in the future.
			logger.Info(entry)

			return nil
		}
	}
}

func newRecover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("recover from err: %v, stack trace:\n%s\n",
						err, debug.Stack())
					handleErr(ctx, NewErr(http.StatusInternalServerError, fmt.Sprintf("%v", err)))
				}
			}()

			return next(ctx)
		}
	}
}

func newErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			err := next(ctx)
			if err != nil {
				method := ctx.Request().Method
				remoteAddr := ctx.Request().RemoteAddr
				path := ctx.Request().URL.Path
				logger.Warnf("[%s] %s - %s request failed: %v", method, remoteAddr, path, err)
				var serviceErr *Err
				if errors.As(err, &serviceErr) {
					handleErr(ctx, serviceErr)
					return err
				}

				handleErr(ctx, NewErr(http.StatusBadRequest, err.Error()))
			}

			return nil
		}
	}
}
