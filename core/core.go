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

package core

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/rest"
	"github.com/LokiWager/analysis-demo/pkg/service"
)

type (
	// DiagnosticConfig is the configuration for diagnostic.
	DiagnosticConfig struct {
		// ProcessID is the process id to diagnose.
		// If it is 0, the current process id will be used.
		// Optional+. Default is 0.
		Pid int
		// EnablePersist is the flag to enable persist the diagnostic data.
		// Default is mongoDB, you can use other storage by implement the interface.
		// Optional+. Default is false.
		EnablePersist bool
		// Port is the port to listen.
		// Optional+. Default is 38080.
		Port int
	}

	// Diagnostic is the diagnostic.
	Diagnostic struct {
		config  *DiagnosticConfig
		Service *service.Service
	}
)

const (
	// DefaultPort is the default port to listen.
	DefaultPort = 38080
)

// NewDiagnostic creates a new diagnostic.
func NewDiagnostic(options *DiagnosticConfig) *Diagnostic {
	return &Diagnostic{
		config: options,
	}
}

// Start starts the diagnostic.
func (d *Diagnostic) Start() {
	logger.Init(&logger.Config{Debug: false})
	if d.config.Pid == 0 {
		d.config.Pid = os.Getpid()
	}

	config := &service.ServiceConfig{
		ProcessID: d.config.Pid,
		Persist:   bool(d.config.EnablePersist),
	}

	port := DefaultPort
	if d.config.Port != 0 {
		port = d.config.Port
	}
	config.ServicePort = port

	go func() {
		logger.Infof("start pprof on port %d", DefaultPort+2)
		err := http.ListenAndServe(fmt.Sprintf(":6060"), nil)
		if err != nil {
			logger.Fatalf("start pprof failed: %v", err)
		}
	}()

	restServer := rest.New(config)
	go restServer.ServerForever(port)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM)

	exit := func() {
		restServer.Shutdown()
	}

	select {
	case <-done:
		logger.Info("!!! RECEIVED THE SIGTERM EXIT SIGNAL, EXITING... !!!")
		exit()
	}

	logger.Info("Graceful Exit Successfully!")
}
