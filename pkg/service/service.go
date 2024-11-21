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

package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v4/process"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/utils/mongodbtool"
)

const (
	DefaultTimeout       = 30
	DefaultTraceFilePath = "./trace"
)

type (
	// ServiceConfig is the configuration for service.
	ServiceConfig struct {
		ProcessID   int  `yaml:"processID" json:"processID"`
		Persist     bool `yaml:"persist" json:"persist"`
		ServicePort int
	}

	// Service is the service.
	Service struct {
		config        *ServiceConfig
		process       *process.Process
		restyClient   *resty.Client
		detectEMA     *EMA
		stopCh        chan struct{}
		traceFilePath string
	}

	// ProcessInfo is the information of a process.
	ProcessInfo struct {
		Name       string   `json:"name"`
		PID        int      `json:"pid"`
		Ppid       int32    `json:"ppid"`
		Cmd        string   `json:"cmd"`
		Pwd        string   `json:"pwd"`
		Envs       []string `json:"envs"`
		NumFDs     int32    `json:"numFDs"`
		NumThreads int32    `json:"numThreads"`
		UserName   string   `json:"userName"`
		Path       string   `json:"path"`
		GoVersion  string   `json:"goVersion"`
	}

	// Connection is the information of a connection.
	Connection struct {
		SourceIP   string `json:"sourceIP"`
		SourcePort int    `json:"sourcePort"`
		DestIP     string `json:"destIP"`
		DestPort   int    `json:"destPort"`
		State      string `json:"state"`
	}
)

// NewService creates a service.
func NewService(config *ServiceConfig) *Service {
	logger.Infof("create service with config pid: %d", config.ProcessID)
	logger.Infof("create service with config persist: %v", config.Persist)

	p, err := process.NewProcess(int32(config.ProcessID))
	if err != nil {
		panic(err)
	}

	client := resty.New().SetTimeout(DefaultTimeout * time.Second)
	// check or mkdir trace file path
	if _, err := os.Stat(DefaultTraceFilePath); os.IsNotExist(err) {
		err = os.Mkdir(DefaultTraceFilePath, 0755)
		if err != nil {
			logger.Fatalf("mkdir trace file path failed: %v", err)
		}
	}
	// walk trace file dir and add all trace files to processTaskMap
	files, err := os.ReadDir(DefaultTraceFilePath)
	if err != nil {
		logger.Fatalf("read trace file dir failed: %v", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		task := &ProcessTask{
			FilePath:  fmt.Sprintf("%s/%s", DefaultTraceFilePath, fileName),
			FileName:  fileName,
			StartTime: time.Now(),
			State:     PendingState,
		}
		processTaskMap.Store(fileName, task)
	}

	return &Service{
		config:        config,
		process:       p,
		restyClient:   client,
		detectEMA:     NewEMA(DefaultAlpha, DefaultThreshold),
		stopCh:        make(chan struct{}),
		traceFilePath: DefaultTraceFilePath,
	}
}

func (s *Service) GetProcessInfo(ctx echo.Context) error {
	name, err := s.process.Name()
	if err != nil {
		logger.Warnf("get process name failed: %v", err)
	}

	ppid, err := s.process.Ppid()
	if err != nil {
		logger.Warnf("get process ppid failed: %v", err)
	}

	cmd, err := s.process.Cmdline()
	if err != nil {
		logger.Warnf("get process cmdline failed: %v", err)
	}

	pwd, err := s.process.Cwd()
	if err != nil {
		logger.Warnf("get process cwd failed: %v", err)
	}

	envs, err := s.process.Environ()
	if err != nil {
		logger.Warnf("get process envs failed: %v", err)
	}

	numFDs, err := s.process.NumFDs()
	if err != nil {
		logger.Warnf("get process numFDs failed: %v", err)
	}

	numThreads, err := s.process.NumThreads()
	if err != nil {
		logger.Warnf("get process numThreads failed: %v", err)
	}

	userName, err := s.process.Username()
	if err != nil {
		logger.Warnf("get process username failed: %v", err)
	}

	path, err := s.process.Exe()
	if err != nil {
		logger.Warnf("get process path failed: %v", err)
	}

	goVersion, err := s.getGoVersion(path)
	if err != nil {
		logger.Warnf("get go version failed: %v", err)
	}

	info := &ProcessInfo{
		Name:       name,
		PID:        s.config.ProcessID,
		Ppid:       ppid,
		Cmd:        cmd,
		Pwd:        pwd,
		Envs:       envs,
		NumFDs:     numFDs,
		NumThreads: numThreads,
		UserName:   userName,
		Path:       path,
		GoVersion:  goVersion,
	}

	return ctx.JSON(200, info)
}

func (s *Service) GetUsage(ctx echo.Context) error {
	if s.config.Persist {
		startTs := ctx.QueryParam("start")
		endTs := ctx.QueryParam("end")
		if endTs == "" {
			endTs = fmt.Sprintf("%d", time.Now().Unix())
		}

		query := bson.M{
			"ts": bson.M{"$gte": startTs, "$lte": endTs},
		}
		if startTs == "" {
			query = bson.M{
				"ts": bson.M{"$lte": endTs},
			}
		}

		metricCollection := mongodbtool.GetCollection("metrics")
		mctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cursor, err := metricCollection.Find(mctx, query)
		if err != nil {
			logger.Warnf("find metrics failed: %v", err)
			return ctx.NoContent(500)
		}
		defer cursor.Close(mctx)

		result := make([]map[string]interface{}, 0)
		for cursor.Next(mctx) {
			var doc map[string]interface{}
			if err := cursor.Decode(&doc); err != nil {
				logger.Warnf("decode metrics failed: %v", err)
				continue
			}
			result = append(result, doc)
		}

		return ctx.JSON(200, result)
	}
	result := s.collectMetrics()
	return ctx.JSON(200, result)
}

func (s *Service) GetOpenFiles(ctx echo.Context) error {
	files, err := s.process.OpenFiles()
	if err != nil {
		logger.Warnf("get process open files failed: %v", err)
	}

	return ctx.JSON(200, files)
}

func (s *Service) GetConnections(ctx echo.Context) error {
	conns, err := s.process.Connections()
	if err != nil {
		logger.Warnf("get process connections failed: %v", err)
	}

	connections := make([]Connection, 0, len(conns))
	for _, conn := range conns {
		connections = append(connections, Connection{
			SourceIP:   conn.Laddr.IP,
			SourcePort: int(conn.Laddr.Port),
			DestIP:     conn.Raddr.IP,
			DestPort:   int(conn.Raddr.Port),
			State:      conn.Status,
		})
	}

	return ctx.JSON(200, connections)
}

func (s *Service) GetProfile(ctx echo.Context) error {
	filePath, fileName, err := s.dumpTraceFile()
	if err != nil {
		logger.Warnf("dump trace file failed: %v", err)
		return ctx.NoContent(500)
	}

	err = s.startTraceTask(fileName, filePath)
	if err != nil {
		logger.Warnf("start trace task failed: %v", err)
		return ctx.NoContent(500)
	}

	return ctx.String(200, fmt.Sprintf("view profile at http://%s.localhost:%d/", fileName, s.config.ServicePort+1))
}

func (s *Service) GetProfileList(ctx echo.Context) error {
	tasks := s.listTraceTasks()
	return ctx.JSON(200, tasks)
}

func (s *Service) StartProfile(ctx echo.Context) error {
	fileName := ctx.QueryParam("file")
	if fileName == "" {
		return ctx.NoContent(400)
	}

	err := s.startTraceTask(fileName, fmt.Sprintf("%s/%s", s.traceFilePath, fileName))
	if err != nil {
		logger.Warnf("start trace task failed: %v", err)
		return ctx.NoContent(500)
	}

	return ctx.String(200, fmt.Sprintf("view profile at http://%s.localhost:%d/", fileName, s.config.ServicePort+1))
}

func (s *Service) StopProfile(ctx echo.Context) error {
	fileName := ctx.QueryParam("file")
	if fileName == "" {
		return ctx.NoContent(400)
	}

	err := s.stopTraceTask(fileName)
	if err != nil {
		logger.Warnf("stop trace task failed: %v", err)
		return ctx.NoContent(500)
	}

	return ctx.NoContent(200)
}

func (s *Service) DeleteProfile(ctx echo.Context) error {
	fileName := ctx.QueryParam("file")
	if fileName == "" {
		return ctx.NoContent(400)
	}

	err := s.deleteTraceFile(fileName)
	if err != nil {
		logger.Warnf("delete trace file failed: %v", err)
		return ctx.NoContent(500)
	}

	return ctx.NoContent(200)
}

func (s *Service) TraceReverseProxy(ctx echo.Context) error {
	// get subdomain
	host := ctx.Request().Header.Get("X-Forwarded-Host")
	if host == "" {
		host = ctx.Request().Host
	}
	ctx.Request().Host = host
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return ctx.NoContent(400)
	}
	subdomain := parts[0]
	err := s.traceProxyHandler(ctx.Response(), ctx.Request(), subdomain)
	if err != nil {
		logger.Warnf("trace proxy handler failed: %v", err)
		return ctx.NoContent(500)
	}

	return nil
}

func (s *Service) GetCustomMetrics(ctx echo.Context) error {
	result := s.collectCustomMetrics()
	return ctx.JSON(200, result)
}

func (s *Service) Close() {
	close(s.stopCh)
}
