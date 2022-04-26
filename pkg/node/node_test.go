package node_test

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
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/logger"
	. "github.com/bhojpur/vpn/pkg/node"
)

var _ = Describe("Node", func() {
	// Trigger key rotation on a low frequency to test everything works in between
	token := GenerateNewConnectionData(25).Base64()

	l := Logger(logger.New(log.LevelFatal))

	Context("Configuration", func() {
		It("fails if is not valid", func() {
			_, err := New(FromBase64(true, true, "  "), WithStore(&blockchain.MemoryStore{}), l)
			Expect(err).To(HaveOccurred())
			_, err = New(FromBase64(true, true, token), WithStore(&blockchain.MemoryStore{}), l)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Connection", func() {
		It("see each other node ID", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			e, _ := New(FromBase64(true, true, token), WithStore(&blockchain.MemoryStore{}), l)
			e2, _ := New(FromBase64(true, true, token), WithStore(&blockchain.MemoryStore{}), l)

			e.Start(ctx)
			e2.Start(ctx)

			Eventually(func() []peer.ID {
				return e.Host().Network().Peers()
			}, 240*time.Second, 1*time.Second).Should(ContainElement(e2.Host().ID()))
		})

		It("nodes can write to the ledger", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			e, _ := New(FromBase64(true, true, token), WithStore(&blockchain.MemoryStore{}), WithDiscoveryInterval(10*time.Second), l)
			e2, _ := New(FromBase64(true, true, token), WithStore(&blockchain.MemoryStore{}), WithDiscoveryInterval(10*time.Second), l)

			e.Start(ctx)
			e2.Start(ctx)

			l, err := e.Ledger()
			Expect(err).ToNot(HaveOccurred())
			l2, err := e2.Ledger()
			Expect(err).ToNot(HaveOccurred())

			l.Announce(ctx, 2*time.Second, func() { l.Add("foo", map[string]interface{}{"bar": "baz"}) })

			Eventually(func() string {
				var s string
				v, exists := l2.GetKey("foo", "bar")
				if exists {
					v.Unmarshal(&s)
				}
				return s
			}, 240*time.Second, 1*time.Second).Should(Equal("baz"))
		})
	})

	Context("connection gater", func() {
		It("blacklists", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			e, _ := New(
				WithBlacklist("1.1.1.1/32", "1.1.1.0/24"),
				FromBase64(true, true, token),
				WithStore(&blockchain.MemoryStore{}),
				l,
			)

			e.Start(ctx)
			addrs := e.ConnectionGater().ListBlockedAddrs()
			peers := e.ConnectionGater().ListBlockedPeers()
			subs := e.ConnectionGater().ListBlockedSubnets()
			Expect(len(addrs)).To(Equal(0))
			Expect(len(peers)).To(Equal(0))
			Expect(len(subs)).To(Equal(2))

			ips := []string{}
			for _, s := range subs {
				ips = append(ips, s.String())
			}
			Expect(ips).To(ContainElements("1.1.1.1/32", "1.1.1.0/24"))
		})
	})
})
