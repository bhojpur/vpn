package process_test

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

	. "github.com/bhojpur/vpn/pkg/utils/processmanager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessManager", func() {
	Context("smoke tests", func() {
		It("starts", func() {
			p := New(
				WithName("/bin/bash"),
				WithArgs("-c", `
while true;
do sleep 1 && echo "test";
done`),
				WithTemporaryStateDir(),
			)
			dir := p.StateDir()
			defer os.RemoveAll(dir)
			Expect(p.Run()).ToNot(HaveOccurred())
			Eventually(func() string {
				c, _ := ioutil.ReadFile(p.StdoutPath())
				return string(c)
			}, "2m").Should(ContainSubstring("test"))

			Expect(p.Stop()).ToNot(HaveOccurred())
		})

		It("stops by reading pid dir", func() {
			dir, err := ioutil.TempDir(os.TempDir(), "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)

			p := New(
				WithStateDir(dir),
				WithName("/bin/bash"),
				WithArgs("-c", `
while true;
do sleep 1 && echo "test";
done`),
			)
			Expect(p.Run()).ToNot(HaveOccurred())

			Eventually(func() string {
				c, _ := ioutil.ReadFile(p.StdoutPath())
				return string(c)
			}, "2m").Should(ContainSubstring("test"))

			Eventually(func() bool {
				return New(
					WithStateDir(dir),
				).IsAlive()
			}, "2m").Should(BeTrue())

			Expect(New(
				WithStateDir(dir)).Stop()).ToNot(HaveOccurred())
		})

	})

	Context("exit codes", func() {
		It("correctly returns 0", func() {
			dir, err := ioutil.TempDir(os.TempDir(), "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)

			p := New(
				WithStateDir(dir),
				WithName("/bin/bash"),
				WithArgs("-c", `
echo "ok"
`),
			)

			Expect(p.Run()).ToNot(HaveOccurred())
			Eventually(func() bool {
				return New(
					WithStateDir(dir),
				).IsAlive()
			}, "2m").Should(BeFalse())
			e, err := New(WithStateDir(dir)).ExitCode()
			Expect(err).ToNot(HaveOccurred())
			Expect(e).To(Equal("0"))
		})
		It("correctly returns 2", func() {
			dir, err := ioutil.TempDir(os.TempDir(), "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)

			p := New(
				WithStateDir(dir),
				WithName("/bin/bash"),
				WithArgs("-c", `
exit 2
`),
			)

			Expect(p.Run()).ToNot(HaveOccurred())
			Eventually(func() bool {
				return New(
					WithStateDir(dir),
				).IsAlive()
			}, "2m").Should(BeFalse())
			e, err := New(WithStateDir(dir)).ExitCode()
			Expect(err).ToNot(HaveOccurred())
			Expect(e).To(Equal("2"))
		})
	})
})
