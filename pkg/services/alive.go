package services

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

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/protocol"
	"github.com/bhojpur/vpn/pkg/utils"
)

func AliveNetworkService(announcetime, scrubTime, maxtime time.Duration) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {
		t := time.Now()
		// By announcing periodically our service to the blockchain
		b.Announce(
			ctx,
			announcetime,
			func() {
				// Keep-alive
				b.Add(protocol.HealthCheckKey, map[string]interface{}{
					n.Host().ID().String(): time.Now().UTC().Format(time.RFC3339),
				})

				// Keep-alive scrub
				nodes := AvailableNodes(b, maxtime)
				if len(nodes) == 0 {
					return
				}
				lead := utils.Leader(nodes)
				if !t.Add(scrubTime).After(time.Now()) {
					// Update timer so not-leader do not attempt to delete bucket afterwards
					// prevent cycles
					t = time.Now()

					if lead == n.Host().ID().String() {
						// Automatically scrub after some time passed
						b.DeleteBucket(protocol.HealthCheckKey)
					}
				}
			},
		)
		return nil
	}
}

// Alive announce the node every announce time, with a periodic scrub time for healthchecks
// the maxtime is the time used to determine when a node is unreachable (after maxtime, its unreachable)
func Alive(announcetime, scrubTime, maxtime time.Duration) []node.Option {
	return []node.Option{
		node.WithNetworkService(AliveNetworkService(announcetime, scrubTime, maxtime)),
	}
}

// AvailableNodes returns the available nodes which sent a healthcheck in the last maxTime
func AvailableNodes(b *blockchain.Ledger, maxTime time.Duration) (active []string) {
	for u, t := range b.LastBlock().Storage[protocol.HealthCheckKey] {
		var s string
		t.Unmarshal(&s)
		parsed, _ := time.Parse(time.RFC3339, s)
		if parsed.Add(maxTime).After(time.Now().UTC()) {
			active = append(active, u)
		}
	}

	return active
}
