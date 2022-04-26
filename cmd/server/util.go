package cmd

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
	"encoding/json"
	"runtime"
	"time"

	"github.com/bhojpur/vpn/pkg/config"
	nodeConfig "github.com/bhojpur/vpn/pkg/config"
	internal "github.com/bhojpur/vpn/pkg/version"
	"github.com/ipfs/go-log"

	vpn "github.com/bhojpur/vpn/pkg/engine"
	"github.com/bhojpur/vpn/pkg/logger"
	node "github.com/bhojpur/vpn/pkg/node"
	"github.com/urfave/cli"
)

var CommonFlags []cli.Flag = []cli.Flag{
	&cli.StringFlag{
		Name:   "config",
		Usage:  "Specify a path to a Bhojpur VPN config file",
		EnvVar: "BHOJPUR_VPN_CONFIG",
	},
	&cli.StringFlag{
		Name:   "timeout",
		Usage:  "Specify a default timeout for connection stream",
		EnvVar: "BHOJPUR_VPN_TIMEOUT",
		Value:  "15s",
	},
	&cli.IntFlag{
		Name:   "mtu",
		Usage:  "Specify a mtu",
		EnvVar: "BHOJPUR_VPN_MTU",
		Value:  1200,
	},
	&cli.BoolTFlag{
		Name:   "bootstrap-iface",
		Usage:  "Setup interface on startup (need privileges)",
		EnvVar: "BHOJPUR_VPN_BOOTSTRAP_IFACE",
	},
	&cli.IntFlag{
		Name:   "packet-mtu",
		Usage:  "Specify a mtu",
		EnvVar: "BHOJPUR_VPN_PACKET_MTU",
		Value:  1420,
	},
	&cli.IntFlag{
		Name:   "channel-buffer-size",
		Usage:  "Specify a channel buffer size",
		EnvVar: "BHOJPUR_VPN_CHANNEL_BUFFERSIZE",
		Value:  0,
	},
	&cli.IntFlag{
		Name:   "discovery-interval",
		Usage:  "DHT discovery interval time",
		EnvVar: "BHOJPUR_VPN_DHT_INTERVAL",
		Value:  120,
	},
	&cli.IntFlag{
		Name:   "ledger-announce-interval",
		Usage:  "Ledger announce interval time",
		EnvVar: "BHOJPUR_VPN_LEDGER_INTERVAL",
		Value:  10,
	},
	&cli.StringFlag{
		Name:   "autorelay-discovery-interval",
		Usage:  "Autorelay discovery interval (Experimental. 0 to disable)",
		EnvVar: "BHOJPUR_VPN_AUTORELAY_DISCOVERY_INTERVAL",
		Value:  "0",
	},
	&cli.IntFlag{
		Name:   "ledger-syncronization-interval",
		Usage:  "Ledger syncronization interval time",
		EnvVar: "BHOJPUR_VPN_LEDGER_SYNC_INTERVAL",
		Value:  10,
	},
	&cli.IntFlag{
		Name:   "nat-ratelimit-global",
		Usage:  "Rate limit global requests",
		EnvVar: "BHOJPUR_VPN_NAT_RATELIMIT_GLOBAL",
		Value:  10,
	},
	&cli.IntFlag{
		Name:   "nat-ratelimit-peer",
		Usage:  "Rate limit perr requests",
		EnvVar: "BHOJPUR_VPN_NAT_RATELIMIT_PEER",
		Value:  10,
	},
	&cli.IntFlag{
		Name:   "nat-ratelimit-interval",
		Usage:  "Rate limit interval",
		EnvVar: "BHOJPUR_VPN_NAT_RATELIMIT_INTERVAL",
		Value:  60,
	},
	&cli.BoolTFlag{
		Name:   "nat-ratelimit",
		Usage:  "Changes the default rate limiting configured in helping other peers determine their reachability status",
		EnvVar: "BHOJPUR_VPN_NAT_RATELIMIT",
	},
	&cli.IntFlag{
		Name:   "max-connections",
		Usage:  "Max connections",
		EnvVar: "BHOJPUR_VPN_MAXCONNS",
		Value:  100,
	},
	&cli.StringFlag{
		Name:   "ledger-state",
		Usage:  "Specify a ledger state directory",
		EnvVar: "BHOJPUR_VPN_LEDGERSTATE",
	},
	&cli.BoolTFlag{
		Name:   "mdns",
		Usage:  "Enable mDNS for peer discovery",
		EnvVar: "BHOJPUR_VPN_MDNS",
	},
	&cli.BoolTFlag{
		Name:   "autorelay",
		Usage:  "Automatically act as a relay if the node can accept inbound connections",
		EnvVar: "BHOJPUR_VPN_AUTORELAY",
	},
	&cli.BoolTFlag{
		Name:   "autorelay-v1",
		Usage:  "Enable autorelay v1 circuits",
		EnvVar: "BHOJPUR_VPN_AUTORELAY_V1",
	},
	&cli.IntFlag{
		Name:  "concurrency",
		Usage: "Number of concurrent requests to serve",
		Value: runtime.NumCPU(),
	},
	&cli.BoolTFlag{
		Name:   "holepunch",
		Usage:  "Automatically try holepunching when possible",
		EnvVar: "BHOJPUR_VPN_HOLE_PUNCH",
	},
	&cli.BoolTFlag{
		Name:   "natservice",
		Usage:  "Tries to determine reachability status of nodes",
		EnvVar: "BHOJPUR_VPN_NAT_SERVICE",
	},
	&cli.BoolTFlag{
		Name:   "natmap",
		Usage:  "Tries to open a port in the firewall via upnp",
		EnvVar: "BHOJPUR_VPN_NAT_MAP",
	},
	&cli.BoolTFlag{
		Name:   "dht",
		Usage:  "Enable DHT for peer discovery",
		EnvVar: "BHOJPUR_VPN_DHT",
	},
	&cli.BoolTFlag{
		Name:   "low-profile",
		Usage:  "Enable low profile. Lowers connections usage",
		EnvVar: "BHOJPUR_VPN_LOW_PROFILE",
	},
	&cli.BoolFlag{
		Name:   "mplex-multiplexer",
		Usage:  "Enable mplex multiplexer.",
		EnvVar: "BHOJPUR_VPN_MPLEX",
	},
	&cli.BoolTFlag{
		Name:   "low-profile-vpn",
		Usage:  "Enable low profile on VPN",
		EnvVar: "BHOJPUR_VPN_LOW_PROFILEVPN",
	},
	&cli.IntFlag{
		Name:   "max-streams",
		Usage:  "Number of concurrent streams",
		Value:  100,
		EnvVar: "BHOJPUR_VPN_MAXSTREAMS",
	},
	&cli.StringFlag{
		Name:   "log-level",
		Usage:  "Specify loglevel",
		EnvVar: "BHOJPUR_VPN_LOG_LEVEL",
		Value:  "info",
	},
	&cli.StringFlag{
		Name:   "libp2p-log-level",
		Usage:  "Specify libp2p loglevel",
		EnvVar: "BHOJPUR_VPN_LIBP2P_LOGLEVEL",
		Value:  "fatal",
	},
	&cli.StringSliceFlag{
		Name:   "discovery-bootstrap-peers",
		Usage:  "List of discovery peers to use",
		EnvVar: "BHOJPUR_VPN_BOOTSTRAP_PEERS",
	},
	&cli.StringSliceFlag{
		Name:   "autorelay-static-peer",
		Usage:  "List of autorelay static peers to use",
		EnvVar: "BHOJPUR_VPN_AUTORELAY_PEERS",
	},
	&cli.StringSliceFlag{
		Name:   "blacklist",
		Usage:  "List of peers/cidr to gate",
		EnvVar: "BHOJPUR_VPN_BLACKLIST",
	},
	&cli.StringFlag{
		Name:   "token",
		Usage:  "Specify a Bhojpur VPN token in place of a config file",
		EnvVar: "BHOJPUR_VPN_TOKEN",
	},
	&cli.StringFlag{
		Name:   "limit-file",
		Usage:  "Specify an limit config (json)",
		EnvVar: "LIMITFILE",
	},
	&cli.StringFlag{
		Name:   "limit-scope",
		Usage:  "Specify a limit scope",
		EnvVar: "LIMITSCOPE",
		Value:  "system",
	},
	&cli.BoolFlag{
		Name:   "limit-config",
		Usage:  "Enable inline resource limit configuration",
		EnvVar: "LIMITCONFIG",
	},
	&cli.BoolFlag{
		Name:   "limit-enable",
		Usage:  "Enable resource manager. (Experimental) All options prefixed with limit requires resource manager to be enabled",
		EnvVar: "LIMITENABLE",
	},
	&cli.BoolFlag{
		Name:   "limit-config-dynamic",
		Usage:  "Enable dynamic resource limit configuration",
		EnvVar: "LIMITCONFIGDYNAMIC",
	},
	&cli.Int64Flag{
		Name:   "limit-config-memory",
		Usage:  "Memory resource limit configuration",
		EnvVar: "LIMITCONFIGMEMORY",
		Value:  128,
	},
	&cli.Float64Flag{
		Name:   "limit-config-memory-fraction",
		Usage:  "Fraction memory resource limit configuration (dynamic)",
		EnvVar: "LIMITCONFIGMEMORYFRACTION",
		Value:  10,
	},
	&cli.Int64Flag{
		Name:   "limit-config-min-memory",
		Usage:  "Minimum memory resource limit configuration (dynamic)",
		EnvVar: "LIMITCONFIGMINMEMORY",
		Value:  10,
	},
	&cli.Int64Flag{
		Name:   "limit-config-max-memory",
		Usage:  "Maximum memory resource limit configuration (dynamic)",
		EnvVar: "LIMITCONFIGMAXMEMORY",
		Value:  200,
	},
	&cli.IntFlag{
		Name:   "limit-config-streams",
		Usage:  "Streams resource limit configuration",
		EnvVar: "LIMITCONFIGSTREAMS",
		Value:  200,
	},
	&cli.IntFlag{
		Name:   "limit-config-streams-inbound",
		Usage:  "Inbound streams resource limit configuration",
		EnvVar: "LIMITCONFIGSTREAMSINBOUND",
		Value:  30,
	},
	&cli.IntFlag{
		Name:   "limit-config-streams-outbound",
		Usage:  "Outbound streams resource limit configuration",
		EnvVar: "LIMITCONFIGSTREAMSOUTBOUND",
		Value:  30,
	},
	&cli.IntFlag{
		Name:   "limit-config-conn",
		Usage:  "Connections resource limit configuration",
		EnvVar: "LIMITCONFIGCONNS",
		Value:  200,
	},
	&cli.IntFlag{
		Name:   "limit-config-conn-inbound",
		Usage:  "Inbound connections resource limit configuration",
		EnvVar: "LIMITCONFIGCONNSINBOUND",
		Value:  30,
	},
	&cli.IntFlag{
		Name:   "limit-config-conn-outbound",
		Usage:  "Outbound connections resource limit configuration",
		EnvVar: "LIMITCONFIGCONNSOUTBOUND",
		Value:  30,
	},
	&cli.IntFlag{
		Name:   "limit-config-fd",
		Usage:  "Max fd resource limit configuration",
		EnvVar: "LIMITCONFIGFD",
		Value:  30,
	},
	&cli.BoolFlag{
		Name:   "peerguard",
		Usage:  "Enable peerguard. (Experimental)",
		EnvVar: "PEERGUARD",
	},
	&cli.BoolFlag{
		Name:   "peergate",
		Usage:  "Enable peergating. (Experimental)",
		EnvVar: "PEERGATE",
	},
	&cli.BoolFlag{
		Name:   "peergate-autoclean",
		Usage:  "Enable peergating autoclean. (Experimental)",
		EnvVar: "PEERGATE_AUTOCLEAN",
	},
	&cli.BoolFlag{
		Name:   "peergate-relaxed",
		Usage:  "Enable peergating relaxation. (Experimental)",
		EnvVar: "PEERGATE_RELAXED",
	},
	&cli.StringFlag{
		Name:   "peergate-auth",
		Usage:  "Peergate auth",
		EnvVar: "PEERGATE_AUTH",
		Value:  "",
	},
	&cli.IntFlag{
		Name:   "peergate-interval",
		Usage:  "Peergater interval time",
		EnvVar: "BHOJPUR_VPN_PEERGATE_INTERVAL",
		Value:  120,
	},
}

func displayStart(ll *logger.Logger) {
	ll.Info(Copyright)

	ll.Infof("Version: %s commit: %s", internal.Version, internal.Commit)
}

func cliToOpts(c *cli.Context) ([]node.Option, []vpn.Option, *logger.Logger) {

	var limitConfig *node.NetLimitConfig

	autorelayInterval, err := time.ParseDuration(c.String("autorelay-discovery-interval"))
	if err != nil {
		autorelayInterval = 0
	}

	if c.Bool("limit-config") {
		limitConfig = &node.NetLimitConfig{
			Dynamic:         c.Bool("limit-config-dynamic"),
			Memory:          c.Int64("limit-config-memory"),
			MinMemory:       c.Int64("limit-config-min-memory"),
			MaxMemory:       c.Int64("limit-config-max-memory"),
			MemoryFraction:  c.Float64("limit-config-memory-fraction"),
			Streams:         c.Int("limit-config-streams"),
			StreamsInbound:  c.Int("limit-config-streams-inbound"),
			StreamsOutbound: c.Int("limit-config-streams-outbound"),
			Conns:           c.Int("limit-config-conn"),
			ConnsInbound:    c.Int("limit-config-conn-inbound"),
			ConnsOutbound:   c.Int("limit-config-conn-outbound"),
			FD:              c.Int("limit-config-fd"),
		}
	}

	// Authproviders are supposed to be passed as a json object
	pa := c.String("peergate-auth")
	d := map[string]map[string]interface{}{}
	json.Unmarshal([]byte(pa), &d)

	nc := nodeConfig.Config{
		NetworkConfig:     c.String("config"),
		NetworkToken:      c.String("token"),
		Address:           c.String("address"),
		Router:            c.String("router"),
		Interface:         c.String("interface"),
		Libp2pLogLevel:    c.String("libp2p-log-level"),
		LogLevel:          c.String("log-level"),
		LowProfile:        c.Bool("low-profile"),
		VPNLowProfile:     c.Bool("low-profile-vpn"),
		Blacklist:         c.StringSlice("blacklist"),
		Concurrency:       c.Int("concurrency"),
		FrameTimeout:      c.String("timeout"),
		ChannelBufferSize: c.Int("channel-buffer-size"),
		InterfaceMTU:      c.Int("mtu"),
		PacketMTU:         c.Int("packet-mtu"),
		BootstrapIface:    c.Bool("bootstrap-iface"),
		Ledger: config.Ledger{
			StateDir:         c.String("ledger-state"),
			AnnounceInterval: time.Duration(c.Int("ledger-announce-interval")) * time.Second,
			SyncInterval:     time.Duration(c.Int("ledger-syncronization-interval")) * time.Second,
		},
		NAT: config.NAT{
			Service:           c.Bool("natservice"),
			Map:               c.Bool("natmap"),
			RateLimit:         c.Bool("nat-ratelimit"),
			RateLimitGlobal:   c.Int("nat-ratelimit-global"),
			RateLimitPeer:     c.Int("nat-ratelimit-peer"),
			RateLimitInterval: time.Duration(c.Int("nat-ratelimit-interval")) * time.Second,
		},
		Discovery: config.Discovery{
			BootstrapPeers: c.StringSlice("discovery-bootstrap-peers"),
			DHT:            c.Bool("dht"),
			MDNS:           c.Bool("mdns"),
			Interval:       time.Duration(c.Int("discovery-interval")) * time.Second,
		},
		Connection: config.Connection{
			AutoRelay:                  c.Bool("autorelay"),
			RelayV1:                    c.Bool("autorelay-v1"),
			MaxConnections:             c.Int("max-connections"),
			MaxStreams:                 c.Int("max-streams"),
			HolePunch:                  c.Bool("holepunch"),
			Mplex:                      c.Bool("mplex-multiplexer"),
			StaticRelays:               c.StringSlice("autorelay-static-peer"),
			AutoRelayDiscoveryInterval: autorelayInterval,
		},
		Limit: config.ResourceLimit{
			Enable:      c.Bool("limit-enable"),
			FileLimit:   c.String("limit-file"),
			Scope:       c.String("limit-scope"),
			MaxConns:    c.Int("max-connections"), // Turn to 0 to use other way of limiting. Files take precedence
			LimitConfig: limitConfig,
		},
		PeerGuard: config.PeerGuard{
			Enable:        c.Bool("peerguard"),
			PeerGate:      c.Bool("peergate"),
			Relaxed:       c.Bool("peergate-relaxed"),
			Autocleanup:   c.Bool("peergate-autoclean"),
			SyncInterval:  time.Duration(c.Int("peergate-interval")) * time.Second,
			AuthProviders: d,
		},
	}

	lvl, err := log.LevelFromString(nc.LogLevel)
	if err != nil {
		lvl = log.LevelError
	}
	llger := logger.New(lvl)

	nodeOpts, vpnOpts, err := nc.ToOpts(llger)
	if err != nil {
		llger.Fatal(err.Error())
	}

	return nodeOpts, vpnOpts, llger
}
