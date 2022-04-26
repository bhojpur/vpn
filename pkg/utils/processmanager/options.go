package process

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
	"io/ioutil"
	"os"
)

// WithKillSignal sets the given signal while attemping to stop. Defaults to "9"
func WithKillSignal(s string) func(cfg *Config) error {
	return func(cfg *Config) error {
		cfg.KillSignal = s
		return nil
	}
}

func WithEnvironment(s ...string) Option {
	return func(cfg *Config) error {
		cfg.Environment = s
		return nil
	}
}

func WithTemporaryStateDir() func(cfg *Config) error {
	return func(cfg *Config) error {
		dir, err := ioutil.TempDir(os.TempDir(), "processmanager")
		cfg.StateDir = dir
		return err
	}
}

func WithStateDir(s string) func(cfg *Config) error {
	return func(cfg *Config) error {
		cfg.StateDir = s
		return nil
	}
}

func WithName(s string) func(cfg *Config) error {
	return func(cfg *Config) error {
		cfg.Name = s
		return nil
	}
}

func WithArgs(s ...string) func(cfg *Config) error {
	return func(cfg *Config) error {
		cfg.Args = append(cfg.Args, s...)
		return nil
	}
}
