package trustzone

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
	"sync"
	"time"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/protocol"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerGater struct {
	sync.Mutex
	trustDB          []peer.ID
	enabled, relaxed bool
}

// NewPeerGater returns a new peergater
// In relaxed mode won't gate until the trustDB contains some auth data.
func NewPeerGater(relaxed bool) *PeerGater {
	return &PeerGater{enabled: true, relaxed: relaxed}
}

// Enabled returns true if the PeerGater is enabled
func (pg *PeerGater) Enabled() bool {
	pg.Lock()
	defer pg.Unlock()
	return pg.enabled
}

// Disables turn off the peer gating mechanism
func (pg *PeerGater) Disable() {
	pg.Lock()
	defer pg.Unlock()
	pg.enabled = false
}

// Enable turns on peer gating mechanism
func (pg *PeerGater) Enable() {
	pg.Lock()
	defer pg.Unlock()
	pg.enabled = true
}

// Implements peergating interface
// resolves to peers in the trustDB. if peer is absent will return true
func (pg *PeerGater) Gate(n *node.Node, p peer.ID) bool {
	pg.Lock()
	defer pg.Unlock()
	if !pg.enabled {
		return false
	}

	if pg.relaxed && len(pg.trustDB) == 0 {
		return false
	}

	for _, pp := range pg.trustDB {
		if pp == p {
			return false
		}
	}

	return true
}

// UpdaterService is a service responsible to sync back trustDB from the ledger state.
// It is a network service which retrieves the senders ID listed in the Trusted Zone
// and fills it in the trustDB used to gate blockchain messages
func (pg *PeerGater) UpdaterService(duration time.Duration) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {
		b.Announce(ctx, duration, func() {
			db := []peer.ID{}
			tz, found := b.CurrentData()[protocol.TrustZoneKey]
			if found {
				for k, _ := range tz {
					db = append(db, peer.ID(k))
				}
			}
			pg.Lock()
			pg.trustDB = db
			pg.Unlock()
		})

		return nil
	}
}
