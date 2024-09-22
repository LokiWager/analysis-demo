/*
 * Copyright (c) 2017, MegaEase
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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/LokiWager/analysis-demo/pkg/engine"
)

var versionTag string
var versionGitCommit string
var versionBuildTime string

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	version := fmt.Sprintf("%s %s.%s", versionTag, versionGitCommit, versionBuildTime)
	logrus.Infof("Version: %s\n", version)

	app := &cli.App{
		Name:    "analysis",
		Usage:   "A CLI tool to analyze Go code",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path", Value: ".", Usage: "Path to the Go source code"},
		},
		Action: func(c *cli.Context) error {
			path := c.String("path")
			if path == "" {
				path = "."
			}

			logrus.Infof("Analyzing Go source code in %s", path)

			// read the go source code
			if _, err := os.Stat(path); os.IsNotExist(err) {
				logrus.Warnf("Path %s does not exist", path)
				os.Exit(1)
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				logrus.Warnf("Failed to read directory %s: %v", path, err)
				os.Exit(1)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				if filepath.Ext(entry.Name()) != ".go" {
					continue
				}

				logrus.Infof("Analyzing %s", entry.Name())
				e := engine.NewEngine(filepath.Join(path, entry.Name()), nil)
				targetIdentifiers, ok := e.CheckIdentifiers()
				if ok {
					for _, targetIdentifier := range targetIdentifiers {
						logrus.Infof("\t Found target identifier \"%s\" at %v", targetIdentifier.Name, targetIdentifier.Pos)
					}
				}

				controlFlow := e.CheckControlFlow()
				if controlFlow {
					logrus.Infof("\t Found 4 level nested control flow")
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
