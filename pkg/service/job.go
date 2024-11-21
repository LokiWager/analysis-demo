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
	"time"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/utils/mongodbtool"
)

const (
	// DefaultAnomalyDetectInterval is the default interval to detect anomaly.
	DefaultAnomalyDetectInterval = 30 * time.Second

	// DefaultThreshold is the default threshold to detect anomaly.
	DefaultThreshold = 1.5

	// DefaultSuspendTimes is the default times to suspend when dump trace file.
	DefaultSuspendTimes = 10
)

func (s *Service) detectAnomaly() {
	ticker := time.NewTicker(DefaultAnomalyDetectInterval)
	defer ticker.Stop()
	// suspending time when dump trace file
	suspendTimes := 0

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if suspendTimes > 0 {
				suspendTimes--
				continue
			}
			cpuPercent, err := s.process.CPUPercent()
			if err != nil {
				logger.Warnf("get process cpu percent failed: %v", err)
				continue
			}
			memPercent, err := s.process.MemoryPercent()
			if err != nil {
				logger.Warnf("get process memory percent failed: %v", err)
				continue
			}
			connCount, err := s.process.Connections()
			if err != nil {
				logger.Warnf("get process connections failed: %v", err)
				continue
			}
			value := EMAValue{
				CPUPercent:    cpuPercent,
				MemoryPercent: float64(memPercent),
				Connections:   float64(len(connCount)),
			}
			s.detectEMA.Update(value)
			if s.detectEMA.IsAnomaly(value) {
				filePath, fileName, err := s.dumpTraceFile()
				if err != nil {
					logger.Warnf("dump trace file failed: %v", err)
				} else {
					suspendTimes = DefaultSuspendTimes
					task := &ProcessTask{
						FilePath:  filePath,
						FileName:  fileName,
						StartTime: time.Now(),
						State:     PendingState,
					}
					processTaskMap.Store(fileName, task)
				}
			}
		}
	}
}

func (s *Service) saveMetrics() {
	ticker := time.NewTicker(DefaultAnomalyDetectInterval)
	defer ticker.Stop()
	metricCollection := mongodbtool.GetCollection("metrics")

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			result := s.collectMetrics()
			result["ts"] = time.Now().Unix()
			_, err := metricCollection.InsertOne(ctx, result)
			if err != nil {
				logger.Warnf("save metrics failed: %v", err)
			}
			cancel()
		}
	}
}
