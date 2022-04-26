package node

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
	"crypto/rand"
	"io"
	mrand "math/rand"
	"net"

	internalCrypto "github.com/bhojpur/vpn/pkg/crypto"

	hub "github.com/bhojpur/vpn/pkg/hub"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	conngater "github.com/libp2p/go-libp2p/p2p/net/conngater"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/xlzd/gotp"
)

// Host returns the libp2p peer host
func (e *Node) Host() host.Host {
	return e.host
}

// ConnectionGater returns the underlying libp2p conngater
func (e *Node) ConnectionGater() *conngater.BasicConnectionGater {
	return e.cg
}

// BlockSubnet blocks the CIDR subnet from connections
func (e *Node) BlockSubnet(cidr string) error {
	// Avoid to loopback traffic by trying to connect to nodes in via VPN
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	return e.ConnectionGater().BlockSubnet(n)
}

func (e *Node) genHost(ctx context.Context) (host.Host, error) {
	var r io.Reader
	if e.seed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(e.seed))
	}

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 4096, r)
	if err != nil {
		return nil, err
	}

	opts := e.config.Options

	cg, err := conngater.NewBasicConnectionGater(nil)
	if err != nil {
		return nil, err
	}

	e.cg = cg

	if e.config.InterfaceAddress != "" {
		e.BlockSubnet(e.config.InterfaceAddress)
	}

	for _, b := range e.config.Blacklist {
		_, net, err := net.ParseCIDR(b)
		if err != nil {
			// Assume it's a peerID
			cg.BlockPeer(peer.ID(b))
		}
		if net != nil {
			cg.BlockSubnet(net)
		}
	}

	opts = append(opts, libp2p.ConnectionGater(cg), libp2p.Identity(prvKey))

	addrs := []multiaddr.Multiaddr{}
	for _, l := range e.config.ListenAddresses {
		addrs = append(addrs, []multiaddr.Multiaddr(l)...)
	}
	opts = append(opts, libp2p.ListenAddrs(addrs...))

	for _, d := range e.config.ServiceDiscovery {
		opts = append(opts, d.Option(ctx))
	}

	opts = append(opts, e.config.AdditionalOptions...)

	if e.config.Insecure {
		e.config.Logger.Info("Disabling Security transport layer")
		opts = append(opts, libp2p.NoSecurity)
	}

	opts = append(opts, libp2p.FallbackDefaults)

	return libp2p.New(opts...)
}

func (e *Node) sealkey() string {
	return internalCrypto.MD5(gotp.NewTOTP(e.config.ExchangeKey, e.config.SealKeyLength, e.config.SealKeyInterval, nil).Now())
}

func (e *Node) handleEvents(ctx context.Context, inputChannel chan *hub.Message, roomMessages chan *hub.Message, pub func(*hub.Message) error, handlers []Handler, peerGater bool) {
	for {
		select {
		case m := <-inputChannel:
			if m == nil {
				continue
			}
			c := m.Copy()
			str, err := e.config.Sealer.Seal(c.Message, e.sealkey())
			if err != nil {
				e.config.Logger.Warnf("%w from %s", err.Error(), c.SenderID)
			}
			c.Message = str

			if err := pub(c); err != nil {
				e.config.Logger.Warnf("publish error: %s", err)
			}

		case m := <-roomMessages:
			if m == nil {
				continue
			}

			if peerGater {
				if e.config.PeerGater != nil && e.config.PeerGater.Gate(e, peer.ID(m.SenderID)) {
					e.config.Logger.Warnf("gated message from %s", m.SenderID)
					continue
				}
			}

			c := m.Copy()
			str, err := e.config.Sealer.Unseal(c.Message, e.sealkey())
			if err != nil {
				e.config.Logger.Warnf("%w from %s", err.Error(), c.SenderID)
			}
			c.Message = str
			e.handleReceivedMessage(c, handlers, inputChannel)
		case <-ctx.Done():
			return
		}
	}
}

func (e *Node) handleReceivedMessage(m *hub.Message, handlers []Handler, c chan *hub.Message) {
	for _, h := range handlers {
		if err := h(e.ledger, m, c); err != nil {
			e.config.Logger.Warnf("handler error: %s", err)
		}
	}
}
