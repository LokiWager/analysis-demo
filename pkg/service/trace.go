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
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/utils/mongodbtool"
)

const (
	PendingState = "pending"
	RunningState = "running"
)

type (
	ProcessTask struct {
		Command   *exec.Cmd `json:"-"`
		FilePath  string    `json:"-"`
		StartTime time.Time `json:"startTime"`
		Port      int       `json:"-"`
		PID       int       `json:"-"`
		FileName  string    `json:"fileName"`
		State     string    `json:"state"`
	}
)

var (
	processTaskMap sync.Map
	portBase       = 32001
	portMutex      sync.Mutex
)

func (s *Service) startTraceTask(fileName, filePath string) error {
	if fileName == "" {
		return fmt.Errorf("file name is empty")
	}
	// check if the task is already running
	if t, ok := processTaskMap.Load(fileName); ok {
		task := t.(*ProcessTask)
		if task.State == RunningState {
			return fmt.Errorf("task is already running, please access it via port %d", task.Port)
		}
	}

	port := s.getAvailablePort()

	if _, err := exec.LookPath("go"); err != nil {
		logger.Warnf("go command not found: %v", err)
		return err
	}

	// create command
	cmd := exec.Command("go", "tool", "trace", "-http=:"+strconv.Itoa(port), filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	if err != nil {
		logger.Warnf("start trace task failed: %v", err)
		return err
	}

	// save task
	task := &ProcessTask{
		Command:   cmd,
		FilePath:  filePath,
		StartTime: time.Now(),
		Port:      port,
		PID:       cmd.Process.Pid,
		FileName:  fileName,
		State:     RunningState,
	}
	processTaskMap.Store(fileName, task)

	return nil
}

func (s *Service) stopTraceTask(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("file name is empty")
	}

	task, ok := processTaskMap.Load(fileName)
	if !ok {
		return fmt.Errorf("task not found")
	}

	err := syscall.Kill(-task.(*ProcessTask).Command.Process.Pid, syscall.SIGKILL)
	if err != nil {
		logger.Warnf("kill trace task failed: %v", err)
		return err
	}
	task.(*ProcessTask).State = PendingState
	processTaskMap.Store(fileName, task)

	return nil
}

func (s *Service) deleteTraceFile(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("file name is empty")
	}

	task, ok := processTaskMap.LoadAndDelete(fileName)
	if !ok {
		return fmt.Errorf("task not found")
	}

	// try to kill the process if it's still running
	if task.(*ProcessTask).State == RunningState && task.(*ProcessTask).Command.Process != nil {
		_ = syscall.Kill(-task.(*ProcessTask).Command.Process.Pid, syscall.SIGKILL)
	}

	err := os.Remove(task.(*ProcessTask).FilePath)
	if err != nil {
		logger.Warnf("delete trace file failed: %v", err)
		return err
	}

	return nil
}

func (s *Service) listTraceTasks() []*ProcessTask {
	tasks := make([]*ProcessTask, 0)
	processTaskMap.Range(func(key, value interface{}) bool {
		tasks = append(tasks, value.(*ProcessTask))
		return true
	})
	return tasks
}

func (s *Service) traceProxyHandler(w http.ResponseWriter, r *http.Request, fileName string) error {
	if fileName == "" {
		return fmt.Errorf("file name is empty")
	}

	taskRaw, ok := processTaskMap.Load(fileName)
	if !ok {
		return fmt.Errorf("task not found")
	}
	task := taskRaw.(*ProcessTask)

	// proxy request
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("%s:%d", "localhost", task.Port)
	}
	var proxyError error = nil
	errorHandler := func(resp http.ResponseWriter, req *http.Request, err error) {
		proxyError = err
		logger.Warnf("http: proxy error: %v", err)
		resp.WriteHeader(http.StatusBadGateway)
	}
	reverseProxy := &httputil.ReverseProxy{Director: director, ErrorHandler: errorHandler}
	myTransport := http.DefaultTransport.(*http.Transport).Clone()
	myTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	reverseProxy.Transport = myTransport
	reverseProxy.ServeHTTP(w, r)
	return proxyError
}

func (s *Service) getAvailablePort() int {
	portMutex.Lock()
	defer portMutex.Unlock()

	port := portBase
	portBase++
	return port
}

func (s *Service) dumpTraceFile() (string, string, error) {
	// create file to trace dir with name trace-<date>.out
	fileName := fmt.Sprintf("trace-%d", time.Now().UnixMilli())
	filePath := fmt.Sprintf("%s/%s", s.traceFilePath, fileName)
	tmpFile, err := os.Create(filePath)
	if err != nil {
		logger.Warnf("create tmp file failed: %v", err)
		return "", "", err
	}

	// write to file
	resp, err := s.restyClient.R().
		SetQueryParam("seconds", "5").
		Execute("GET", fmt.Sprintf("http://localhost:6060/debug/pprof/trace"))
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		logger.Warnf("get profile failed: %v", err)
		return "", "", err
	}

	traceData := resp.Body()
	err = os.WriteFile(filePath, traceData, 0644)
	if err != nil {
		logger.Warnf("write trace data to file failed: %v", err)
		return "", "", err
	}

	return filePath, fileName, nil
}

func (s *Service) dumpTraceFileToMDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	traceCollection := mongodbtool.GetCollection("trace")

	// write to file
	resp, err := s.restyClient.R().
		SetQueryParam("seconds", "5").
		Execute("GET", "http://localhost:6060/debug/pprof/trace")
	if err != nil {
		logger.Warnf("get profile failed: %v", err)
		return err
	}

	traceData := resp.Body()
	doc := bson.M{
		"trace": traceData,
		"date":  time.Now(),
	}
	_, err = traceCollection.InsertOne(ctx, doc)
	if err != nil {
		logger.Warnf("insert trace data to mongodb failed: %v", err)
		return err
	}

	return nil
}
