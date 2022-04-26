package logger

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"os"

	terminal "github.com/bhojpur/vpn/pkg/utils/terminal"

	"github.com/ipfs/go-log"
	"github.com/pterm/pterm"
)

var _ log.StandardLogger = &Logger{}

type Logger struct {
	level log.LogLevel
}

func New(lvl log.LogLevel) *Logger {
	if !terminal.IsTerminal(os.Stdout) {
		pterm.DisableColor()
	}
	if lvl == log.LevelDebug {
		pterm.EnableDebugMessages()
	}
	return &Logger{level: lvl}
}

func joinMsg(args ...interface{}) (message string) {
	for _, m := range args {
		message += " " + fmt.Sprintf("%v", m)
	}
	return
}

func (l Logger) enabled(lvl log.LogLevel) bool {
	return lvl >= l.level
}

func (l Logger) Debug(args ...interface{}) {
	if l.enabled(log.LevelDebug) {
		pterm.Debug.Println(joinMsg(args...))
	}
}

func (l Logger) Debugf(f string, args ...interface{}) {
	if l.enabled(log.LevelDebug) {
		pterm.Debug.Printfln(f, args...)
	}
}

func (l Logger) Error(args ...interface{}) {
	if l.enabled(log.LevelError) {
		pterm.Error.Println(pterm.LightRed(joinMsg(args...)))
	}
}

func (l Logger) Errorf(f string, args ...interface{}) {
	if l.enabled(log.LevelError) {
		pterm.Error.Printfln(pterm.LightRed(f), args...)
	}
}

func (l Logger) Fatal(args ...interface{}) {
	if l.enabled(log.LevelFatal) {
		pterm.Fatal.Println(pterm.Red(joinMsg(args...)))
	}
}

func (l Logger) Fatalf(f string, args ...interface{}) {
	if l.enabled(log.LevelFatal) {
		pterm.Fatal.Printfln(pterm.Red(f), args...)
	}
}

func (l Logger) Info(args ...interface{}) {
	if l.enabled(log.LevelInfo) {
		pterm.Info.Println(pterm.LightBlue(joinMsg(args...)))
	}
}

func (l Logger) Infof(f string, args ...interface{}) {
	if l.enabled(log.LevelInfo) {
		pterm.Info.Printfln(pterm.LightBlue(f), args...)
	}
}

func (l Logger) Panic(args ...interface{}) {
	l.Fatal(args...)
}

func (l Logger) Panicf(f string, args ...interface{}) {
	l.Fatalf(f, args...)
}

func (l Logger) Warn(args ...interface{}) {
	if l.enabled(log.LevelWarn) {
		pterm.Warning.Println(pterm.LightYellow(joinMsg(args...)))
	}
}

func (l Logger) Warnf(f string, args ...interface{}) {
	if l.enabled(log.LevelWarn) {
		pterm.Warning.Printfln(pterm.LightYellow(f), args...)
	}
}

func (l Logger) Warning(args ...interface{}) {
	l.Warn(args...)
}

func (l Logger) Warningf(f string, args ...interface{}) {
	l.Warnf(f, args...)
}
