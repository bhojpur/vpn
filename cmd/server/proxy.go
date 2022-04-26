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

	"github.com/bhojpur/vpn/pkg/api"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/services"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/urfave/cli"
)

func Proxy() cli.Command {
	return cli.Command{
		Name:        "proxy",
		Usage:       "Starts a local HTTP proxy server to egress nodes",
		Description: `Start a proxy locally, providing an ingress point for the network.`,
		UsageText:   "vpnsvr proxy",
		Flags: append(CommonFlags,
			&cli.StringFlag{
				Name:   "listen",
				Value:  ":8080",
				Usage:  "Listening address",
				EnvVar: "PROXYLISTEN",
			},
			&cli.BoolFlag{
				Name: "debug",
			},
			&cli.IntFlag{
				Name:   "interval",
				Usage:  "proxy announce time interval",
				EnvVar: "PROXYINTERVAL",
				Value:  120,
			},
			&cli.IntFlag{
				Name:   "dead-interval",
				Usage:  "interval (in seconds) wether detect egress nodes offline",
				EnvVar: "PROXYDEADINTERVAL",
				Value:  600,
			},
		),
		Action: func(c *cli.Context) error {
			o, _, ll := cliToOpts(c)

			o = append(o, services.Proxy(
				time.Duration(c.Int("interval"))*time.Second,
				time.Duration(c.Int("dead-interval"))*time.Second,
				c.String("listen"))...)

			bwc := metrics.NewBandwidthCounter()
			o = append(o, node.WithLibp2pAdditionalOptions(libp2p.BandwidthReporter(bwc)))

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

			return api.API(ctx, c.String("listen"), 5*time.Second, 20*time.Second, e, bwc, c.Bool("debug"))
		},
	}
}
