package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/LokiWager/analysis-demo/pkg/service"
	"net/http"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/labstack/echo/v4"
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
	s.group.GET("/process/usages", s.service.GetUsage)
}

func (s *Server) ServerForever(port int) {
	err := s.app.Start(fmt.Sprintf(":%d", port))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
	logger.Infof("api server listening on %d", port)
}

// Shutdown closes the Server.
func (s *Server) Shutdown() {
	err := s.app.Shutdown(context.Background())
	if err != nil {
		logger.Errorf("shutdown api server failed: %v", err)
	}
}
