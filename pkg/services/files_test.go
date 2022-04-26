package services_test

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
	"time"

	"github.com/ipfs/go-log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/logger"
	node "github.com/bhojpur/vpn/pkg/node"
	. "github.com/bhojpur/vpn/pkg/services"
)

var _ = Describe("File services", func() {
	token := node.GenerateNewConnectionData(25).Base64()

	logg := logger.New(log.LevelError)
	l := node.Logger(logg)

	e2, _ := node.New(
		node.WithDiscoveryInterval(10*time.Second),
		node.WithNetworkService(AliveNetworkService(2*time.Second, 4*time.Second, 15*time.Minute)),
		node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)

	Context("File sharing", func() {
		It("sends and receive files between two nodes", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fileUUID := "test"

			f, err := ioutil.TempFile("", "test")
			Expect(err).ToNot(HaveOccurred())

			defer os.RemoveAll(f.Name())

			ioutil.WriteFile(f.Name(), []byte("testfile"), os.ModePerm)

			// First node expose a file
			opts, err := ShareFile(logg, 10*time.Second, fileUUID, f.Name())
			Expect(err).ToNot(HaveOccurred())

			opts = append(opts, node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)
			e, _ := node.New(opts...)

			e.Start(ctx)
			e2.Start(ctx)

			Eventually(func() string {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				f, err := ioutil.TempFile("", "test")
				Expect(err).ToNot(HaveOccurred())

				defer os.RemoveAll(f.Name())

				ll, _ := e2.Ledger()
				ll1, _ := e.Ledger()
				By(fmt.Sprint(ll.CurrentData(), ll.LastBlock().Index, ll1.CurrentData()))
				ReceiveFile(ctx, ll, e2, logg, 2*time.Second, fileUUID, f.Name())
				b, _ := ioutil.ReadFile(f.Name())
				return string(b)
			}, 190*time.Second, 1*time.Second).Should(Equal("testfile"))
		})
	})
})
