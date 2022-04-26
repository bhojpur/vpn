package service

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
	"os/exec"
	"path/filepath"

	process "github.com/bhojpur/vpn/pkg/utils/processmanager"
)

// NewProcessController returns a new process controller associated with the state directory
func NewProcessController(statedir string) *ProcessController {
	return &ProcessController{stateDir: statedir}
}

// ProcessController syntax sugar around go-processmanager
type ProcessController struct {
	stateDir string
}

// Process returns a process associated within binaries inside the state dir
func (a *ProcessController) Process(state, p string, opts ...process.Option) *process.Process {
	return process.New(
		append(opts,
			process.WithName(a.BinaryPath(p)),
			process.WithStateDir(filepath.Join(a.stateDir, "proc", state)),
		)...,
	)
}

// BinaryPath returns the binary path of the program requested as argument.
// The binary path is relative to the process state directory
func (a *ProcessController) BinaryPath(b string) string {
	return filepath.Join(a.stateDir, "bin", b)
}

// Run simply runs a command from a binary in the state directory
func (a *ProcessController) Run(command string, args ...string) (string, error) {
	cmd := exec.Command(a.BinaryPath(command), args...)
	out, err := cmd.CombinedOutput()

	return string(out), err
}
