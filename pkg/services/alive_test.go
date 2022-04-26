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
	"time"

	"github.com/ipfs/go-log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/logger"
	node "github.com/bhojpur/vpn/pkg/node"
	. "github.com/bhojpur/vpn/pkg/services"
)

var _ = Describe("Alive service", func() {
	token := node.GenerateNewConnectionData().Base64()

	logg := logger.New(log.LevelError)
	l := node.Logger(logg)

	opts := append(
		Alive(5*time.Second, 100*time.Second, 15*time.Minute),
		node.WithDiscoveryInterval(10*time.Second),
		node.FromBase64(true, true, token),
		l)

	Context("Aliveness check", func() {
		It("detect both nodes alive after a while", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			e2, _ := node.New(append(opts, node.WithStore(&blockchain.MemoryStore{}))...)
			e1, _ := node.New(append(opts, node.WithStore(&blockchain.MemoryStore{}))...)

			e1.Start(ctx)
			e2.Start(ctx)

			ll, _ := e1.Ledger()

			ll.Persist(ctx, 5*time.Second, 100*time.Second, "t", "t", "test")

			matches := And(ContainElement(e2.Host().ID().String()),
				ContainElement(e1.Host().ID().String()))

			index := ll.LastBlock().Index
			Eventually(func() []string {
				ll, err := e1.Ledger()
				if err != nil {
					return []string{}
				}
				return AvailableNodes(ll, 15*time.Minute)
			}, 100*time.Second, 1*time.Second).Should(matches)

			Expect(ll.LastBlock().Index).ToNot(Equal(index))
		})
	})

	Context("Aliveness Scrub", func() {
		BeforeEach(func() {
			opts = append(
				Alive(10*time.Second, 30*time.Second, 15*time.Minute),
				node.WithDiscoveryInterval(10*time.Second),
				node.FromBase64(true, true, token),
				l)
		})

		It("cleans up after a while", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			e2, _ := node.New(append(opts, node.WithStore(&blockchain.MemoryStore{}))...)
			e1, _ := node.New(append(opts, node.WithStore(&blockchain.MemoryStore{}))...)

			e1.Start(ctx)
			time.Sleep(5 * time.Second)
			e2.Start(ctx)

			ll, _ := e1.Ledger()

			ll.Persist(ctx, 5*time.Second, 100*time.Second, "t", "t", "test")

			matches := And(ContainElement(e2.Host().ID().String()),
				ContainElement(e1.Host().ID().String()))

			index := ll.LastBlock().Index
			Eventually(func() []string {
				ll, err := e1.Ledger()
				if err != nil {
					return []string{}
				}
				return AvailableNodes(ll, 15*time.Minute)
			}, 120*time.Second, 1*time.Second).Should(matches)

			Expect(ll.LastBlock().Index).ToNot(Equal(index))
			index = ll.LastBlock().Index

			Eventually(func() []string {
				ll, err := e1.Ledger()
				if err != nil {
					return []string{}
				}
				return AvailableNodes(ll, 15*time.Minute)
			}, 360*time.Second, 1*time.Second).Should(BeEmpty())

			Expect(ll.LastBlock().Index).ToNot(Equal(index))
			index = ll.LastBlock().Index

			Eventually(func() []string {
				ll, err := e1.Ledger()
				if err != nil {
					return []string{}
				}
				return AvailableNodes(ll, 15*time.Minute)
			}, 60*time.Second, 1*time.Second).Should(matches)
			Expect(ll.LastBlock().Index).ToNot(Equal(index))

		})
	})
})
