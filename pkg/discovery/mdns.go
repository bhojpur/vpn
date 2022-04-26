package discovery

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
	"github.com/libp2p/go-libp2p"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns_legacy"
)

type MDNS struct {
	DiscoveryServiceTag string
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
	c log.StandardLogger
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	//n.c.Infof("mDNS: discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		n.c.Debugf("mDNS: error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

func (d *MDNS) Option(ctx context.Context) func(c *libp2p.Config) error {
	return func(*libp2p.Config) error { return nil }
}

func (d *MDNS) Run(l log.StandardLogger, ctx context.Context, host host.Host) error {
	// setup mDNS discovery to find local peers
	// XXX: Valid for new mdns
	// disc := mdns.NewMdnsService(host, d.DiscoveryServiceTag, &discoveryNotifee{h: host, c: l})
	// return disc.Start()
	// We stick to legacy atm as mdns 0.15 is kinda of broken
	// see: https://github.com/libp2p/go-libp2p/pull/1192
	disc, err := mdns.NewMdnsService(ctx, host, time.Hour, d.DiscoveryServiceTag)
	if err != nil {
		return err
	}

	n := &discoveryNotifee{h: host, c: l}

	disc.RegisterNotifee(n)

	return nil
}
