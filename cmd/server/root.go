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
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"

	"github.com/bhojpur/vpn/pkg/api"
	vpn "github.com/bhojpur/vpn/pkg/engine"
	"github.com/bhojpur/vpn/pkg/node"
	bvpn "github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/services"
	"github.com/urfave/cli"
)

const Copyright string = `Copyright (c) 2018 Bhojpur Consulting Private Limited, India.`

func MainFlags() []cli.Flag {
	basedir, _ := os.UserHomeDir()
	if basedir == "" {
		basedir = os.TempDir()
	}

	return append([]cli.Flag{
		&cli.IntFlag{
			Name:  "key-otp-interval",
			Usage: "Tweaks default otp interval (in seconds) when generating new tokens",
			Value: 9000,
		},
		&cli.BoolFlag{
			Name:  "g",
			Usage: "Generates a new configuration and prints it on screen",
		},
		&cli.BoolFlag{
			Name:  "b",
			Usage: "Encodes the new config in base64, so it can be used as a token",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Starts API with pprof attached",
		},
		&cli.BoolFlag{
			Name:   "api",
			Usage:  "Starts also the API daemon locally for inspecting the network status",
			EnvVar: "API",
		},
		&cli.StringFlag{
			Name:   "api-listen",
			Value:  ":8080",
			Usage:  "API listening port",
			EnvVar: "APILISTEN",
		},
		&cli.BoolFlag{
			Name:   "dhcp",
			Usage:  "Enables p2p ip negotiation (experimental)",
			EnvVar: "DHCP",
		},
		&cli.BoolFlag{
			Name:   "transient-conn",
			Usage:  "Allow transient connections",
			EnvVar: "TRANSIENTCONN",
		},
		&cli.StringFlag{
			Name:   "lease-dir",
			Value:  filepath.Join(basedir, ".bhojpur", "leases"),
			Usage:  "DHCP leases directory",
			EnvVar: "DHCPLEASEDIR",
		},
		&cli.StringFlag{
			Name:   "address",
			Usage:  "VPN virtual address",
			EnvVar: "ADDRESS",
			Value:  "10.1.0.1/24",
		},
		&cli.StringFlag{
			Name:   "dns",
			Usage:  "DNS listening address. Empty to disable dns server",
			EnvVar: "DNSADDRESS",
			Value:  "",
		},
		&cli.BoolTFlag{
			Name:   "dns-forwarder",
			Usage:  "Enables dns forwarding",
			EnvVar: "DNSFORWARD",
		},
		&cli.BoolFlag{
			Name:   "egress",
			Usage:  "Enables nodes for egress",
			EnvVar: "EGRESS",
		},
		&cli.IntFlag{
			Name:   "egress-announce-time",
			Usage:  "Egress announce time (s)",
			EnvVar: "EGRESSANNOUNCE",
			Value:  200,
		},
		&cli.IntFlag{
			Name:   "dns-cache-size",
			Usage:  "DNS LRU cache size",
			EnvVar: "DNSCACHESIZE",
			Value:  200,
		},
		&cli.IntFlag{
			Name:   "aliveness-healthcheck-interval",
			Usage:  "Healthcheck interval",
			EnvVar: "HEALTHCHECKINTERVAL",
			Value:  120,
		},
		&cli.IntFlag{
			Name:   "aliveness-healthcheck-scrub-interval",
			Usage:  "Healthcheck scrub interval",
			EnvVar: "HEALTHCHECKSCRUBINTERVAL",
			Value:  600,
		},
		&cli.IntFlag{
			Name:   "aliveness-healthcheck-max-interval",
			Usage:  "Healthcheck max interval. Threshold after a node is determined offline",
			EnvVar: "HEALTHCHECKMAXINTERVAL",
			Value:  900,
		},
		&cli.StringSliceFlag{
			Name:   "dns-forward-server",
			Usage:  "List of DNS forward server, e.g. 8.8.8.8:53, 192.168.1.1:53 ...",
			EnvVar: "DNSFORWARDSERVER",
			Value:  &cli.StringSlice{"8.8.8.8:53", "1.1.1.1:53"},
		},
		&cli.StringFlag{
			Name:   "router",
			Usage:  "Sends all packets to this node",
			EnvVar: "ROUTER",
		},
		&cli.StringFlag{
			Name:   "interface",
			Usage:  "Interface name",
			Value:  "bhojpurvpn0",
			EnvVar: "IFACE",
		}}, CommonFlags...)
}

func Main() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Bool("g") {
			// Generates a new config and exit
			newData := bvpn.GenerateNewConnectionData(c.Int("key-otp-interval"))
			if c.Bool("b") {
				fmt.Print(newData.Base64())
			} else {
				fmt.Println(newData.YAML())
			}

			os.Exit(0)
		}
		o, vpnOpts, ll := cliToOpts(c)

		// Egress and DHCP needs the Alive service
		// DHCP needs alive services enabled to all nodes, also those with a static IP.
		o = append(o,
			services.Alive(
				time.Duration(c.Int("aliveness-healthcheck-interval"))*time.Second,
				time.Duration(c.Int("aliveness-healthcheck-scrub-interval"))*time.Second,
				time.Duration(c.Int("aliveness-healthcheck-max-interval"))*time.Second)...)

		if c.Bool("dhcp") {
			// Adds DHCP server
			address, _, err := net.ParseCIDR(c.String("address"))
			if err != nil {
				return err
			}
			nodeOpts, vO := vpn.DHCP(ll, 15*time.Minute, c.String("lease-dir"), address.String())
			o = append(o, nodeOpts...)
			vpnOpts = append(vpnOpts, vO...)
		}

		if c.Bool("egress") {
			o = append(o, services.Egress(time.Duration(c.Int("egress-announce-time"))*time.Second)...)
		}

		dns := c.String("dns")
		if dns != "" {
			// Adds DNS Server
			o = append(o,
				services.DNS(ll, dns,
					c.Bool("dns-forwarder"),
					c.StringSlice("dns-forward-server"),
					c.Int("dns-cache-size"),
				)...)
		}

		bwc := metrics.NewBandwidthCounter()
		if c.Bool("api") {
			o = append(o, node.WithLibp2pAdditionalOptions(libp2p.BandwidthReporter(bwc)))
		}

		opts, err := vpn.Register(vpnOpts...)
		if err != nil {
			return err
		}

		e, err := bvpn.New(append(o, opts...)...)
		if err != nil {
			return err
		}

		displayStart(ll)

		ctx := context.Background()

		if c.Bool("transient-conn") {
			ctx = network.WithUseTransient(ctx, "accept")
		}

		if c.Bool("api") {
			go api.API(ctx, c.String("api-listen"), 5*time.Second, 20*time.Second, e, bwc, c.Bool("debug"))
		}

		return e.Start(ctx)
	}
}
