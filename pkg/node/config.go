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
	"time"

	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/bhojpur/vpn/pkg/blockchain"
	discovery "github.com/bhojpur/vpn/pkg/discovery"
	hub "github.com/bhojpur/vpn/pkg/hub"
	protocol "github.com/bhojpur/vpn/pkg/protocol"
)

// Config is the node configuration
type Config struct {
	// ExchangeKey is a Symmetric key used to seal the messages
	ExchangeKey string

	// RoomName is the OTP token gossip room where all peers are subscribed to
	RoomName string

	// ListenAddresses is the discovery peer initial bootstrap addresses
	ListenAddresses []discovery.AddrList

	// Insecure disables secure p2p e2e encrypted communication
	Insecure bool

	// Handlers are a list of handlers subscribed to messages received by the vpn interface
	Handlers, GenericChannelHandler []Handler

	MaxMessageSize  int
	SealKeyInterval int

	ServiceDiscovery []ServiceDiscovery
	NetworkServices  []NetworkService
	Logger           log.StandardLogger

	SealKeyLength    int
	InterfaceAddress string

	Store blockchain.Store

	// Handle is a handle consumed by HumanInterfaces to handle received messages
	Handle                     func(bool, *hub.Message)
	StreamHandlers             map[protocol.Protocol]StreamHandler
	AdditionalOptions, Options []libp2p.Option

	DiscoveryInterval, LedgerSyncronizationTime, LedgerAnnounceTime time.Duration
	DiscoveryBootstrapPeers                                         discovery.AddrList

	Whitelist, Blacklist []string

	// GenericHub enables generic hub
	GenericHub bool

	Sealer    Sealer
	PeerGater Gater
}

type Gater interface {
	Gate(*Node, peer.ID) bool
	Enable()
	Disable()
	Enabled() bool
}

type Sealer interface {
	Seal(string, string) (string, error)
	Unseal(string, string) (string, error)
}

// NetworkService is a service running over the network. It takes a context, a node and a ledger
type NetworkService func(context.Context, Config, *Node, *blockchain.Ledger) error

type StreamHandler func(*Node, *blockchain.Ledger) func(stream network.Stream)

type Handler func(*blockchain.Ledger, *hub.Message, chan *hub.Message) error

type ServiceDiscovery interface {
	Run(log.StandardLogger, context.Context, host.Host) error
	Option(context.Context) func(c *libp2p.Config) error
}

type Option func(cfg *Config) error

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}
