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
	"time"

	client "github.com/bhojpur/vpn/pkg/api/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bhojpur/vpn/pkg/api/client/service"
)

var _ = Describe("Service", func() {
	c := client.NewClient(client.WithHost(testInstance))
	s := NewClient("foo", c)
	Context("Retrieves nodes", func() {
		PIt("Detect nodes", func() {
			Eventually(func() []string {
				n, _ := s.ActiveNodes()
				return n
			},
				100*time.Second, 1*time.Second).ShouldNot(BeEmpty())
		})
	})

	Context("Advertize nodes", func() {
		It("Detect nodes", func() {
			n, err := s.AdvertizingNodes()
			Expect(len(n)).To(Equal(0))
			Expect(err).ToNot(HaveOccurred())

			s.Advertize("foo")

			Eventually(func() []string {
				n, _ := s.AdvertizingNodes()
				return n
			},
				100*time.Second, 1*time.Second).Should(Equal([]string{"foo"}))
		})
	})
})
