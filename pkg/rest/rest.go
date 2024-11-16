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
		service *service.Service
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
		service: svc,
	}

	s.setupAPIs()

	return s
}

func (s *Server) setupAPIs() {
	s.group.GET("/process/info", s.service.GetProcessInfo)
	s.group.GET("/process/fds", s.service.GetOpenFiles)
	s.group.GET("/process/usage", s.service.GetUsage)
	s.group.GET("/process/connections", s.service.GetConnections)
	s.group.GET("/process/generate-profile", s.service.GetProfile)
	s.group.GET("/process/profiles", s.service.GetProfileList)
	s.group.GET("/process/start-profile", s.service.StartProfile)
	s.group.GET("/process/stop-profile", s.service.StopProfile)
	s.group.DELETE("/process/delete-profile", s.service.DeleteProfile)
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

		proxy.Any("/*", s.service.TraceReverseProxy)

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
