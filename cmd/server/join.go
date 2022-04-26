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

	"github.com/bhojpur/vpn/pkg/node"
	"github.com/urfave/cli"
)

func Start() cli.Command {
	return cli.Command{
		Name:  "start",
		Usage: "Start the network without activating any interface",
		Description: `Connect over the P2P network without establishing a VPN.
Useful for setting up relays or hop nodes to improve the network connectivity.`,
		UsageText: "vpnsvr start",
		Flags:     CommonFlags,
		Action: func(c *cli.Context) error {
			o, _, ll := cliToOpts(c)
			e, err := node.New(o...)
			if err != nil {
				return err
			}

			displayStart(ll)

			// Start the node to the network, using our ledger
			if err := e.Start(context.Background()); err != nil {
				return err
			}

			ll.Info("Joining p2p network")
			<-context.Background().Done()
			return nil
		},
	}
}
