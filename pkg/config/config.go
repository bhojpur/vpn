package config

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
	"fmt"
	"math/bits"
	"os"
	"strings"
	"time"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/crypto"
	"github.com/bhojpur/vpn/pkg/discovery"
	vpn "github.com/bhojpur/vpn/pkg/engine"
	"github.com/bhojpur/vpn/pkg/logger"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/services"
	"github.com/bhojpur/vpn/pkg/trustzone"
	"github.com/bhojpur/vpn/pkg/trustzone/authprovider/ecdsa"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	connmanager "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	rcmgr "github.com/libp2p/go-libp2p-resource-manager"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/peterbourgon/diskv"
	"github.com/songgao/water"
)

// Config is the config struct for the node and the default Bhojpur VPN services
// It is used to generate opts for the node and the services before start.
type Config struct {
	NetworkConfig, NetworkToken                string
	Address                                    string
	Router                                     string
	Interface                                  string
	Libp2pLogLevel, LogLevel                   string
	LowProfile, VPNLowProfile, BootstrapIface  bool
	Blacklist                                  []string
	Concurrency                                int
	FrameTimeout                               string
	ChannelBufferSize, InterfaceMTU, PacketMTU int
	NAT                                        NAT
	Connection                                 Connection
	Discovery                                  Discovery
	Ledger                                     Ledger
	Limit                                      ResourceLimit
	// PeerGuard (experimental)
	// enable peerguardian and add specific auth options
	PeerGuard PeerGuard
}

type PeerGuard struct {
	Enable      bool
	Relaxed     bool
	Autocleanup bool
	PeerGate    bool
	// AuthProviders in the freemap form:
	// ecdsa:
	//   private_key: "foo_bar"
	AuthProviders map[string]map[string]interface{}
	SyncInterval  time.Duration
}

type ResourceLimit struct {
	FileLimit   string
	LimitConfig *node.NetLimitConfig
	Scope       string
	MaxConns    int
	StaticMin   int64
	StaticMax   int64
	Enable      bool
}

// Ledger is the ledger configuration structure
type Ledger struct {
	AnnounceInterval, SyncInterval time.Duration
	StateDir                       string
}

// Discovery allows to enable/disable discovery and
// set bootstrap peers
type Discovery struct {
	DHT, MDNS      bool
	BootstrapPeers []string
	Interval       time.Duration
}

// Connection is the configuration section
// relative to the connection services
type Connection struct {
	HolePunch bool
	AutoRelay bool

	AutoRelayDiscoveryInterval time.Duration
	RelayV1                    bool
	StaticRelays               []string

	Mplex          bool
	MaxConnections int
	MaxStreams     int
}

// NAT is the structure relative to NAT configuration settings
// It allows to enable/disable the service and NAT mapping, and rate limiting too.
type NAT struct {
	Service   bool
	Map       bool
	RateLimit bool

	RateLimitGlobal, RateLimitPeer int
	RateLimitInterval              time.Duration
}

// Validate returns error if the configuration is not valid
func (c Config) Validate() error {
	if c.NetworkConfig == "" &&
		c.NetworkToken == "" {
		return fmt.Errorf("BHOJPUR_VPN_CONFIG or BHOJPUR_VPN_TOKEN not supplied. At least a config file is required")
	}
	return nil
}

func peers2List(peers []string) discovery.AddrList {
	addrsList := discovery.AddrList{}
	for _, p := range peers {
		addrsList.Set(p)
	}
	return addrsList
}
func peers2AddrInfo(peers []string) []peer.AddrInfo {
	addrsList := []peer.AddrInfo{}
	for _, p := range peers {
		pi, err := peer.AddrInfoFromString(p)
		if err == nil {
			addrsList = append(addrsList, *pi)
		}

	}
	return addrsList
}

// ToOpts returns node and vpn options from a configuration
func (c Config) ToOpts(l *logger.Logger) ([]node.Option, []vpn.Option, error) {

	if err := c.Validate(); err != nil {
		return nil, nil, err
	}

	config := c.NetworkConfig
	address := c.Address
	router := c.Router
	iface := c.Interface
	logLevel := c.LogLevel
	libp2plogLevel := c.Libp2pLogLevel
	dhtE, mDNS := c.Discovery.DHT, c.Discovery.MDNS

	ledgerState := c.Ledger.StateDir

	peers := c.Discovery.BootstrapPeers

	lvl, err := log.LevelFromString(logLevel)
	if err != nil {
		lvl = log.LevelError
	}

	llger := logger.New(lvl)

	libp2plvl, err := log.LevelFromString(libp2plogLevel)
	if err != nil {
		libp2plvl = log.LevelFatal
	}

	token := c.NetworkToken

	addrsList := peers2List(peers)

	dhtOpts := []dht.Option{}

	if c.LowProfile {
		dhtOpts = append(dhtOpts, dht.BucketSize(20))
	}

	opts := []node.Option{
		node.WithDiscoveryInterval(c.Discovery.Interval),
		node.WithLedgerAnnounceTime(c.Ledger.AnnounceInterval),
		node.WithLedgerInterval(c.Ledger.SyncInterval),
		node.Logger(llger),
		node.WithDiscoveryBootstrapPeers(addrsList),
		node.WithBlacklist(c.Blacklist...),
		node.LibP2PLogLevel(libp2plvl),
		node.WithInterfaceAddress(address),
		node.WithSealer(&crypto.AESSealer{}),
		node.FromBase64(mDNS, dhtE, token, dhtOpts...),
		node.FromYaml(mDNS, dhtE, config, dhtOpts...),
	}

	vpnOpts := []vpn.Option{
		vpn.WithConcurrency(c.Concurrency),
		vpn.WithInterfaceAddress(address),
		vpn.WithLedgerAnnounceTime(c.Ledger.AnnounceInterval),
		vpn.Logger(llger),
		vpn.WithTimeout(c.FrameTimeout),
		vpn.WithInterfaceType(water.TUN),
		vpn.NetLinkBootstrap(c.BootstrapIface),
		vpn.WithChannelBufferSize(c.ChannelBufferSize),
		vpn.WithInterfaceMTU(c.InterfaceMTU),
		vpn.WithPacketMTU(c.PacketMTU),
		vpn.WithRouterAddress(router),
		vpn.WithInterfaceName(iface),
		vpn.WithMaxStreams(c.Connection.MaxStreams),
	}

	if c.VPNLowProfile {
		vpnOpts = append(vpnOpts, vpn.LowProfile)
	}

	libp2pOpts := []libp2p.Option{libp2p.UserAgent("BhojpurVPN")}

	// AutoRelay section configuration
	if c.Connection.AutoRelay {
		relayOpts := []autorelay.Option{}
		if c.Connection.RelayV1 {
			relayOpts = append(relayOpts, autorelay.WithCircuitV1Support())
		}

		// If no relays are specified and no discovery interval, then just use default static relays (to be deprecated)
		if len(c.Connection.StaticRelays) == 0 && c.Connection.AutoRelayDiscoveryInterval == 0 {
			relayOpts = append(relayOpts, autorelay.WithDefaultStaticRelays())
		} else if len(c.Connection.StaticRelays) > 0 {
			relayOpts = append(relayOpts, autorelay.WithStaticRelays(peers2AddrInfo(c.Connection.StaticRelays)))
		} else {
			peerChan := make(chan peer.AddrInfo)
			// Add AutoRelayFeederService (needs a DHT Service discovery)
			opts = append(opts, func(cfg *node.Config) error {
				for _, sd := range cfg.ServiceDiscovery {
					switch d := sd.(type) {
					case *discovery.DHT:
						llger.Debugf("DHT automatic relay discovery configured every '%s'\n", c.Connection.AutoRelayDiscoveryInterval.String())
						cfg.NetworkServices = append(cfg.NetworkServices, services.AutoRelayFeederService(llger, peerChan, d, c.Connection.AutoRelayDiscoveryInterval))
					}
				}
				return nil
			})
			relayOpts = append(relayOpts, autorelay.WithPeerSource(peerChan))
		}

		libp2pOpts = append(libp2pOpts,
			libp2p.EnableAutoRelay(relayOpts...))
	}

	if c.Connection.Mplex {
		libp2pOpts = append(libp2pOpts,
			libp2p.ChainOptions(
				libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
				libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
			))
	}

	if c.NAT.RateLimit {
		libp2pOpts = append(libp2pOpts, libp2p.AutoNATServiceRateLimit(
			c.NAT.RateLimitGlobal,
			c.NAT.RateLimitPeer,
			c.NAT.RateLimitInterval,
		))
	}

	if c.Connection.MaxConnections != 0 {
		cm, err := connmanager.NewConnManager(
			1,
			c.Connection.MaxConnections,
			connmanager.WithGracePeriod(80*time.Second),
		)
		if err != nil {
			llger.Fatal("could not create connection manager")
		}

		libp2pOpts = append(libp2pOpts, libp2p.ConnectionManager(cm))
	}

	if !c.Limit.Enable {
		libp2pOpts = append(libp2pOpts, libp2p.ResourceManager(network.NullResourceManager))
	} else {
		var limiter *rcmgr.BasicLimiter

		if c.Limit.FileLimit != "" {
			limitFile, err := os.Open(c.Limit.FileLimit)
			if err != nil {
				return opts, vpnOpts, err
			}
			defer limitFile.Close()

			limiter, err = rcmgr.NewDefaultLimiterFromJSON(limitFile)
			if err != nil {
				return opts, vpnOpts, err
			}
		} else if c.Limit.MaxConns != 0 {
			min := int64(1 << 30)
			max := int64(4 << 30)
			if c.Limit.StaticMin != 0 {
				min = c.Limit.StaticMin
			}
			if c.Limit.StaticMax != 0 {
				max = c.Limit.StaticMax
			}

			defaultLimits := rcmgr.DefaultLimits.WithSystemMemory(.125, min, max)

			maxconns := int(c.Limit.MaxConns)
			if 2*maxconns > defaultLimits.SystemBaseLimit.ConnsInbound {
				// adjust conns to 2x to allow for two conns per peer (TCP+QUIC)
				defaultLimits.SystemBaseLimit.ConnsInbound = logScale(2 * maxconns)
				defaultLimits.SystemBaseLimit.ConnsOutbound = logScale(2 * maxconns)
				defaultLimits.SystemBaseLimit.Conns = logScale(4 * maxconns)

				defaultLimits.SystemBaseLimit.StreamsInbound = logScale(16 * maxconns)
				defaultLimits.SystemBaseLimit.StreamsOutbound = logScale(64 * maxconns)
				defaultLimits.SystemBaseLimit.Streams = logScale(64 * maxconns)

				if 2*maxconns > defaultLimits.SystemBaseLimit.FD {
					defaultLimits.SystemBaseLimit.FD = logScale(2 * maxconns)
				}

				defaultLimits.ServiceBaseLimit.StreamsInbound = logScale(8 * maxconns)
				defaultLimits.ServiceBaseLimit.StreamsOutbound = logScale(32 * maxconns)
				defaultLimits.ServiceBaseLimit.Streams = logScale(32 * maxconns)

				defaultLimits.ProtocolBaseLimit.StreamsInbound = logScale(8 * maxconns)
				defaultLimits.ProtocolBaseLimit.StreamsOutbound = logScale(32 * maxconns)
				defaultLimits.ProtocolBaseLimit.Streams = logScale(32 * maxconns)
			}
			limiter = rcmgr.NewStaticLimiter(defaultLimits)

		} else {
			limiter = rcmgr.NewDefaultLimiter()
		}

		libp2p.SetDefaultServiceLimits(limiter)

		rc, err := rcmgr.NewResourceManager(limiter)
		if err != nil {
			llger.Fatal("could not create resource manager")
		}

		if c.Limit.LimitConfig != nil {
			if err := node.NetSetLimit(rc, c.Limit.Scope, *c.Limit.LimitConfig); err != nil {
				return opts, vpnOpts, err
			}
		}

		libp2pOpts = append(libp2pOpts, libp2p.ResourceManager(rc))
	}

	if c.Connection.HolePunch {
		libp2pOpts = append(libp2pOpts, libp2p.EnableHolePunching())
	}

	if c.NAT.Service {
		libp2pOpts = append(libp2pOpts, libp2p.EnableNATService())
	}

	if c.NAT.Map {
		libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())
	}

	opts = append(opts, node.WithLibp2pOptions(libp2pOpts...))

	if ledgerState != "" {
		opts = append(opts, node.WithStore(blockchain.NewDiskStore(diskv.New(diskv.Options{
			BasePath:     ledgerState,
			CacheSizeMax: uint64(50), // 50MB
		}))))
	} else {
		opts = append(opts, node.WithStore(&blockchain.MemoryStore{}))
	}

	if c.PeerGuard.Enable {
		pg := trustzone.NewPeerGater(c.PeerGuard.Relaxed)
		dur := c.PeerGuard.SyncInterval

		// Build up the authproviders for the peerguardian
		aps := []trustzone.AuthProvider{}
		for ap, providerOpts := range c.PeerGuard.AuthProviders {
			a, err := authProvider(llger, ap, providerOpts)
			if err != nil {
				return opts, vpnOpts, fmt.Errorf("invalid authprovider: %w", err)
			}
			aps = append(aps, a)
		}

		pguardian := trustzone.NewPeerGuardian(llger, aps...)

		opts = append(opts,
			node.WithNetworkService(
				pg.UpdaterService(dur),
				pguardian.Challenger(dur, c.PeerGuard.Autocleanup),
			),
			node.EnableGenericHub,
			node.GenericChannelHandlers(pguardian.ReceiveMessage),
		)
		// We always pass a PeerGater such will be registered to the API if necessary
		opts = append(opts, node.WithPeerGater(pg))
		// IF it's not enabled, we just disable it right away.
		if !c.PeerGuard.PeerGate {
			pg.Disable()
		}
	}

	return opts, vpnOpts, nil
}

func authProvider(ll log.StandardLogger, s string, opts map[string]interface{}) (trustzone.AuthProvider, error) {
	switch strings.ToLower(s) {
	case "ecdsa":
		pk, exists := opts["private_key"]
		if !exists {
			return nil, fmt.Errorf("No private key provided")
		}
		return ecdsa.ECDSA521Provider(ll, fmt.Sprint(pk))
	}
	return nil, fmt.Errorf("not supported")
}

func logScale(val int) int {
	bitlen := bits.Len(uint(val))
	return 1 << bitlen
}
