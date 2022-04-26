package utils_test

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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bhojpur/vpn/pkg/utils"
)

var _ = Describe("Leader utilities", func() {
	Context("Leader", func() {
		It("returns the correct leader", func() {
			Expect(Leader([]string{"a", "b", "c", "d"})).To(Equal("b"))
			Expect(Leader([]string{"a", "b", "c", "d", "e", "f", "G", "bb"})).To(Equal("b"))
			Expect(Leader([]string{"a", "b", "c", "d", "e", "f", "G", "bb", "z", "b1", "b2"})).To(Equal("z"))
			Expect(Leader([]string{"1", "2", "3", "4", "5"})).To(Equal("2"))
			Expect(Leader([]string{"1", "2", "3", "4", "5", "6", "7", "21", "22"})).To(Equal("22"))
		})
	})
})
