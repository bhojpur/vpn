package api_test

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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/bhojpur/vpn/pkg/api"
	client "github.com/bhojpur/vpn/pkg/api/client"
	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/logger"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/ipfs/go-log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API", func() {

	Context("Binds on socket", func() {
		It("sets data to the API", func() {
			d, _ := ioutil.TempDir("", "xxx")
			defer os.RemoveAll(d)
			os.MkdirAll(d, os.ModePerm)
			socket := filepath.Join(d, "socket")

			c := client.NewClient(client.WithHost("unix://" + socket))

			token := node.GenerateNewConnectionData().Base64()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := node.Logger(logger.New(log.LevelFatal))

			e, _ := node.New(node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)
			e.Start(ctx)

			e2, _ := node.New(node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)
			e2.Start(ctx)

			go func() {
				err := API(ctx, fmt.Sprintf("unix://%s", socket), 10*time.Second, 20*time.Second, e, nil, false)
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(func() error {
				return c.Put("b", "f", "bar")
			}, 10*time.Second, 1*time.Second).ShouldNot(HaveOccurred())

			Eventually(c.GetBuckets, 100*time.Second, 1*time.Second).Should(ContainElement("b"))

			Eventually(func() string {
				d, err := c.GetBucketKey("b", "f")
				if err != nil {
					fmt.Println(err)
				}
				var s string

				d.Unmarshal(&s)
				return s
			}, 10*time.Second, 1*time.Second).Should(Equal("bar"))
		})
	})
})
