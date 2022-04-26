package client_test

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
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bhojpur/vpn/pkg/api/client"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

var _ = Describe("Client", func() {
	c := NewClient(WithHost(testInstance))

	Context("Operates blockchain", func() {
		var testBucket string

		AfterEach(func() {
			Eventually(c.GetBuckets, 100*time.Second, 1*time.Second).Should(ContainElement(testBucket))
			err := c.DeleteBucket(testBucket)
			Expect(err).ToNot(HaveOccurred())
			Eventually(c.GetBuckets, 100*time.Second, 1*time.Second).ShouldNot(ContainElement(testBucket))
		})

		BeforeEach(func() {
			testBucket = randStringBytes(10)
		})

		It("Puts string data", func() {
			err := c.Put(testBucket, "foo", "bar")
			Expect(err).ToNot(HaveOccurred())

			Eventually(c.GetBuckets, 100*time.Second, 1*time.Second).Should(ContainElement(testBucket))
			Eventually(func() ([]string, error) { return c.GetBucketKeys(testBucket) }, 100*time.Second, 1*time.Second).Should(ContainElement("foo"))

			Eventually(func() (string, error) {
				resp, err := c.GetBucketKey(testBucket, "foo")
				if err == nil {
					var r string
					resp.Unmarshal(&r)
					return r, nil
				}
				return "", err
			}, 100*time.Second, 1*time.Second).Should(Equal("bar"))

			m, err := c.Ledger()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(m) > 0).To(BeTrue())
		})

		It("Puts random data", func() {
			err := c.Put(testBucket, "foo2", struct{ Foo string }{Foo: "bar"})
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() (string, error) {
				resp, err := c.GetBucketKey(testBucket, "foo2")
				if err == nil {
					var r struct{ Foo string }
					resp.Unmarshal(&r)
					return r.Foo, nil
				}

				return "", err
			}, 100*time.Second, 1*time.Second).Should(Equal("bar"))
		})
	})
})
