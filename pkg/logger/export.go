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

package logger

func Warnf(fmt string, args ...interface{}) {
	compoundSystemLogger.Warnf(fmt, args...)
}

func Debugf(fmt string, args ...interface{}) {
	compoundSystemLogger.Debugf(fmt, args...)
}

func Errorf(fmt string, args ...interface{}) {
	compoundSystemLogger.Errorf(fmt, args...)
}

func Infof(fmt string, args ...interface{}) {
	compoundSystemLogger.Infof(fmt, args...)
}

func Fatalf(fmt string, args ...interface{}) {
	compoundSystemLogger.Fatalf(fmt, args...)
}

func Panicf(fmt string, args ...interface{}) {
	compoundSystemLogger.Panicf(fmt, args...)
}

func Warn(args string) {
	compoundSystemLogger.Warn(args)
}

func Debug(args string) {
	compoundSystemLogger.Debug(args)
}

func Info(args string) {
	compoundSystemLogger.Info(args)
}

func Error(args string) {
	compoundSystemLogger.Error(args)
}

func Fatal(args string) {
	compoundSystemLogger.Fatal(args)
}

func Panic(args string) {
	compoundSystemLogger.Panic(args)
}
