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

package main

import (
	"net/http"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/rest"
	"github.com/LokiWager/analysis-demo/pkg/service"

	_ "net/http/pprof"
)

func main() {
	logger.Init(&logger.Config{Debug: false})
	app := &cli.App{
		Name:  "diagnostic",
		Usage: "diagnostic is a tool to diagnose the running process",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "pid",
				Value: 0,
				Usage: "The process id to diagnose",
			},
			&cli.BoolFlag{
				Name:  "persist",
				Value: false,
				Usage: "The path to persist the diagnostic data",
			},
		},
		Action: func(c *cli.Context) error {
			pid := c.Int("pid")
			if pid == 0 {
				pid = os.Getpid()
			}
			config := &service.ServiceConfig{
				ProcessID: pid,
			}

			if c.Bool("persist") {
				config.Persist = true
			}

			rest.New(config).ServerForever(8080)
			return nil
		},
	}

	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			logger.Fatalf("start pprof failed: %v", err)
		}
	}()

	logger.Info("diagnostic server starting")
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
