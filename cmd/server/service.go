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
	"errors"
	"time"

	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/services"
	"github.com/urfave/cli"
)

func cliNameAddress(c *cli.Context) (name, address string, err error) {
	name = c.Args().Get(0)
	address = c.Args().Get(1)
	if name == "" && c.String("name") == "" {
		err = errors.New("Either a file UUID as first argument or with --name needs to be provided")
		return
	}
	if address == "" && c.String("address") == "" {
		err = errors.New("Either a file UUID as first argument or with --name needs to be provided")
		return
	}
	if c.String("name") != "" {
		name = c.String("name")
	}
	if c.String("address") != "" {
		address = c.String("address")
	}
	return name, address, nil
}

func ServiceAdd() cli.Command {
	return cli.Command{
		Name:    "service-add",
		Aliases: []string{"sa"},
		Usage:   "Expose a service to the network without creating a VPN",
		Description: `Expose a local or a remote endpoint connection as a service in the VPN. 
		The host will act as a proxy between the service and the connection`,
		UsageText: "vpnsvr service-add unique-id ip:port",
		Flags: append(CommonFlags,
			cli.StringFlag{
				Name:  "name",
				Usage: `Unique name of the service to be server over the network.`,
			},
			cli.StringFlag{
				Name: "address",
				Usage: `Remote address that the service is running to. That can be a remote webserver, a local SSH server, etc.
For example, '192.168.1.1:80', or '127.0.0.1:22'.`,
			},
		),
		Action: func(c *cli.Context) error {
			name, address, err := cliNameAddress(c)
			if err != nil {
				return err
			}
			o, _, ll := cliToOpts(c)

			o = append(o, services.RegisterService(ll, time.Duration(c.Int("ledger-announce-interval"))*time.Second, name, address)...)

			e, err := node.New(o...)
			if err != nil {
				return err
			}

			displayStart(ll)

			// Join the node to the network, using our ledger
			if err := e.Start(context.Background()); err != nil {
				return err
			}

			for {
				time.Sleep(2 * time.Second)
			}
		},
	}
}

func ServiceConnect() cli.Command {
	return cli.Command{
		Aliases: []string{"sc"},
		Usage:   "Connects to a service in the network without creating a VPN",
		Name:    "service-connect",
		Description: `Bind a local port to connect to a remote service in the network.
Creates a local listener which connects over the service in the network without creating a VPN.
`,
		UsageText: "vpnsvr service-connect unique-id (ip):port",
		Flags: append(CommonFlags,
			cli.StringFlag{
				Name:  "name",
				Usage: `Unique name of the service in the network.`,
			},
			cli.StringFlag{
				Name: "address",
				Usage: `Address where to bind locally. E.g. ':8080'. A proxy will be created
to the service over the network`,
			},
		),
		Action: func(c *cli.Context) error {
			name, address, err := cliNameAddress(c)
			if err != nil {
				return err
			}
			o, _, ll := cliToOpts(c)
			e, err := node.New(
				append(o,
					node.WithNetworkService(
						services.ConnectNetworkService(
							time.Duration(c.Int("ledger-announce-interval"))*time.Second,
							name,
							address,
						),
					),
				)...,
			)
			if err != nil {
				return err
			}
			displayStart(ll)

			// starts the node
			return e.Start(context.Background())
		},
	}
}
