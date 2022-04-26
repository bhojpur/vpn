package service_test

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
	"testing"
	"time"

	. "github.com/bhojpur/vpn/pkg/api/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testInstance = os.Getenv("TEST_INSTANCE")

func TestService(t *testing.T) {
	if testInstance == "" {
		fmt.Println("a testing instance has to be defined with TEST_INSTANCE")
		os.Exit(1)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

var _ = BeforeSuite(func() {
	// Start the test suite only if we have some machines connected

	Eventually(func() (int, error) {
		c := NewClient(WithHost(testInstance))
		m, err := c.Machines()
		return len(m), err
	}, 100*time.Second, 1*time.Second).Should(BeNumerically(">=", 0))
})
