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
	"sync"
	"time"

	internalCrypto "github.com/bhojpur/vpn/pkg/crypto"

	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/xlzd/gotp"
)

type DHT struct {
	OTPKey               string
	OTPInterval          int
	KeyLength            int
	RendezvousString     string
	BootstrapPeers       AddrList
	latestRendezvous     string
	RefreshDiscoveryTime time.Duration
	*dht.IpfsDHT
	dhtOptions []dht.Option
}

func NewDHT(d ...dht.Option) *DHT {
	return &DHT{dhtOptions: d}
}

func (d *DHT) Option(ctx context.Context) func(c *libp2p.Config) error {
	return libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		// make the DHT with the given Host
		return d.startDHT(ctx, h)
	})
}
func (d *DHT) Rendezvous() string {
	if d.OTPKey != "" {
		totp := gotp.NewTOTP(d.OTPKey, d.KeyLength, d.OTPInterval, nil)

		//totp := gotp.NewDefaultTOTP(d.OTPKey)
		rv := internalCrypto.MD5(totp.Now())
		d.latestRendezvous = rv
		return rv
	}
	return d.RendezvousString
}

func (d *DHT) startDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	if d.IpfsDHT == nil {
		// Start a DHT, for use in peer discovery. We can't just make a new DHT
		// client because we want each peer to maintain its own local copy of the
		// DHT, so that the bootstrapping node of the DHT can go down without
		// inhibiting future peer discovery.
		kad, err := dht.New(ctx, h, d.dhtOptions...)
		if err != nil {
			return d.IpfsDHT, err
		}
		d.IpfsDHT = kad
	}

	return d.IpfsDHT, nil
}

func (d *DHT) Run(c log.StandardLogger, ctx context.Context, host host.Host) error {
	if d.KeyLength == 0 {
		d.KeyLength = 12
	}

	if len(d.BootstrapPeers) == 0 {
		d.BootstrapPeers = dht.DefaultBootstrapPeers
	}
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := d.startDHT(ctx, host)
	if err != nil {
		return err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	c.Info("Bootstrapping DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return err
	}

	connect := func() {
		d.bootstrapPeers(c, ctx, host)
		if d.latestRendezvous != "" {
			d.announceAndConnect(c, ctx, kademliaDHT, host, d.latestRendezvous)
		}

		rv := d.Rendezvous()
		d.announceAndConnect(c, ctx, kademliaDHT, host, rv)
	}

	go func() {
		connect()
		for {
			// We don't want a ticker here but a timer
			// this is less "talkative" as a DHT connect() can take up
			// long time and can exceed d.RefreshdiscoveryTime.
			// In this way we ensure we wait at least timeout to fire a connect()
			timer := time.NewTimer(d.RefreshDiscoveryTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				connect()
			}
		}
	}()

	return nil
}

func (d *DHT) bootstrapPeers(c log.StandardLogger, ctx context.Context, host host.Host) {
	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range d.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if host.Network().Connectedness(peerinfo.ID) != network.Connected {
				if err := host.Connect(ctx, *peerinfo); err != nil {
					c.Debug(err.Error())
				} else {
					c.Debug("Connection established with bootstrap node:", *peerinfo)
				}
			}
		}()
	}
	wg.Wait()
}

func (d *DHT) announceAndConnect(l log.StandardLogger, ctx context.Context, kademliaDHT *dht.IpfsDHT, host host.Host, rv string) error {
	l.Debug("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, rv)
	l.Debug("Successfully announced!")
	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	l.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, rv)
	if err != nil {
		return err
	}

	for p := range peerChan {
		// Don't dial ourselves or peers without address
		if p.ID == host.ID() || len(p.Addrs) == 0 {
			continue
		}

		if host.Network().Connectedness(p.ID) != network.Connected {
			l.Debug("Found peer:", p)
			if err := host.Connect(ctx, p); err != nil {
				l.Debug("Failed connecting to", p)
			} else {
				l.Debug("Connected to:", p)
			}
		} else {
			l.Debug("Known peer (already connected):", p)
		}
	}

	return nil
}
