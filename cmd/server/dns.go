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
	"time"

	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/services"
	"github.com/urfave/cli"
)

func DNS() cli.Command {
	return cli.Command{
		Name:        "dns",
		Usage:       "Starts a local dns server",
		Description: `Start a local dns server which uses the blockchain to resolve addresses`,
		UsageText:   "vpnsvr dns",
		Flags: append(CommonFlags,
			&cli.StringFlag{
				Name:   "listen",
				Usage:  "DNS listening address. Empty to disable dns server",
				EnvVar: "DNSADDRESS",
				Value:  "",
			},
			&cli.BoolTFlag{
				Name:   "dns-forwarder",
				Usage:  "Enables dns forwarding",
				EnvVar: "DNSFORWARD",
			},
			&cli.IntFlag{
				Name:   "dns-cache-size",
				Usage:  "DNS LRU cache size",
				EnvVar: "DNSCACHESIZE",
				Value:  200,
			},
			&cli.StringSliceFlag{
				Name:   "dns-forward-server",
				Usage:  "List of DNS forward server, e.g. 8.8.8.8:53, 192.168.1.1:53 ...",
				EnvVar: "DNSFORWARDSERVER",
				Value:  &cli.StringSlice{"8.8.8.8:53", "1.1.1.1:53"},
			},
		),
		Action: func(c *cli.Context) error {
			o, _, ll := cliToOpts(c)

			dns := c.String("listen")
			// Adds DNS Server
			o = append(o,
				services.DNS(ll, dns,
					c.Bool("dns-forwarder"),
					c.StringSlice("dns-forward-server"),
					c.Int("dns-cache-size"),
				)...)

			e, err := node.New(o...)
			if err != nil {
				return err
			}

			displayStart(ll)

			ctx := context.Background()
			// Start the node to the network, using our ledger
			if err := e.Start(ctx); err != nil {
				return err
			}

			for {
				time.Sleep(1 * time.Second)
			}
		},
	}
}
