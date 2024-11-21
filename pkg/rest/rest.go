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
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/service"
)

type (
	// Server is the HTTP API server.
	Server struct {
		app     *echo.Echo
		group   *echo.Group
		Service *service.Service
	}
)

const (
	// APIPrefix is the common prefix of all apis.
	APIPrefix = "/api/v1"
)

// New creates an API server.
func New(cfg *service.ServiceConfig) *Server {
	app := echo.New()

	app.HideBanner = true
	app.HidePort = true

	app.Use(newLogger())
	app.Use(newRecover())
	app.Use(newErrorHandler())

	serviceGroup := app.Group(APIPrefix)
	svc := service.NewService(cfg)

	s := &Server{
		app:     app,
		group:   serviceGroup,
		Service: svc,
	}

	s.setupAPIs()

	return s
}

func (s *Server) setupAPIs() {
	s.group.GET("/process/info", s.Service.GetProcessInfo)
	s.group.GET("/process/fds", s.Service.GetOpenFiles)
	s.group.GET("/process/usage", s.Service.GetUsage)
	s.group.GET("/process/connections", s.Service.GetConnections)
	s.group.GET("/process/generate-profile", s.Service.GetProfile)
	s.group.GET("/process/profiles", s.Service.GetProfileList)
	s.group.GET("/process/start-profile", s.Service.StartProfile)
	s.group.GET("/process/stop-profile", s.Service.StopProfile)
	s.group.DELETE("/process/delete-profile", s.Service.DeleteProfile)
	s.group.GET("/process/custom-metrics", s.Service.GetCustomMetrics)
}

func (s *Server) ServerForever(port int) {
	go func() {
		logger.Infof("proxy server listening on %d", port+1)
		proxy := echo.New()
		proxy.HideBanner = true
		proxy.HidePort = true

		proxy.Use(newLogger())
		proxy.Use(newRecover())
		proxy.Use(newErrorHandler())

		proxy.Any("/*", s.Service.TraceReverseProxy)

		err := proxy.Start(fmt.Sprintf(":%d", port+1))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	logger.Infof("api server listening on %d", port)
	err := s.app.Start(fmt.Sprintf(":%d", port))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

// Shutdown closes the Server.
func (s *Server) Shutdown() {
	err := s.app.Shutdown(context.Background())
	if err != nil {
		logger.Errorf("shutdown api server failed: %v", err)
	}
}
