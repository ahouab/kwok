/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"golang.org/x/exp/slog"
)

type Level = slog.Level

const (
	DebugLevel Level = slog.DebugLevel
	InfoLevel  Level = slog.InfoLevel
	WarnLevel  Level = slog.WarnLevel
	ErrorLevel Level = slog.ErrorLevel
)

func wrapSlog(log *slog.Logger) *Logger {
	return &Logger{log}
}

type Logger struct {
	log *slog.Logger
}

func (l *Logger) LogDepth(calldepth int, level Level, msg string, args ...any) {
	l.log.LogDepth(calldepth+1, level, msg, args...)
}

func (l *Logger) Log(level Level, msg string, args ...any) {
	l.LogDepth(0, level, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.LogDepth(0, DebugLevel, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.LogDepth(0, InfoLevel, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.LogDepth(0, WarnLevel, msg, args...)
}

func (l *Logger) Error(msg string, err error, args ...any) {
	if err != nil {
		args = append(args[:len(args):len(args)], slog.String(slog.ErrorKey, err.Error()))
	}
	l.log.LogDepth(0, ErrorLevel, msg, args...)
}

func (l *Logger) With(args ...any) *Logger {
	return wrapSlog(l.log.With(args...))
}

func (l *Logger) WithGroup(name string) *Logger {
	return wrapSlog(l.log.WithGroup(name))
}
