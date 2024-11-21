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

const (
	DefaultAlpha = 0.3
)

type (
	// EMA is the Exponential Moving Average.
	EMA struct {
		alpha  float64
		value  EMAValue
		init   bool
		thread float64
	}

	EMAValue struct {
		CPUPercent    float64 `json:"cpu_percent"`
		MemoryPercent float64 `json:"memory_percent"`
		Connections   float64 `json:"connections"`
	}
)

// NewEMA creates an EMA.
func NewEMA(alpha, thread float64) *EMA {
	return &EMA{
		alpha:  alpha,
		thread: thread,
	}
}

// Update updates the EMA.
func (e *EMA) Update(value EMAValue) {
	if !e.init {
		e.value = value
		e.init = true
	} else {
		e.value.CPUPercent = e.alpha*value.CPUPercent + (1-e.alpha)*e.value.CPUPercent
		e.value.MemoryPercent = e.alpha*value.MemoryPercent + (1-e.alpha)*e.value.MemoryPercent
		e.value.Connections = e.alpha*value.Connections + (1-e.alpha)*e.value.Connections
	}
}

// Value returns the value of EMA.
func (e *EMA) Value() EMAValue {
	return e.value
}

// IsAnomaly returns if the value is anomaly.
func (e *EMA) IsAnomaly(value EMAValue) bool {
	return value.CPUPercent > e.value.CPUPercent*e.thread ||
		value.MemoryPercent > e.value.MemoryPercent*e.thread ||
		value.Connections > e.value.Connections*e.thread
}
