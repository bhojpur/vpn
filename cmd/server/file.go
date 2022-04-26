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

func cliNamePath(c *cli.Context) (name, path string, err error) {
	name = c.Args().Get(0)
	path = c.Args().Get(1)
	if name == "" && c.String("name") == "" {
		err = errors.New("Either a file UUID as first argument or with --name needs to be provided")
		return
	}
	if path == "" && c.String("path") == "" {
		err = errors.New("Either a file UUID as first argument or with --name needs to be provided")
		return
	}
	if c.String("name") != "" {
		name = c.String("name")
	}
	if c.String("path") != "" {
		path = c.String("path")
	}
	return name, path, nil
}

func FileSend() cli.Command {
	return cli.Command{
		Name:        "file-send",
		Aliases:     []string{"fs"},
		Usage:       "Serve a file to the network",
		Description: `Serve a file to the network without connecting over VPN`,
		UsageText:   "vpnsvr file-send unique-id /src/path",
		Flags: append(CommonFlags,
			cli.StringFlag{
				Name:     "name",
				Required: true,
				Usage: `Unique name of the file to be served over the network. 
This is also the ID used to refer when receiving it.`,
			},
			cli.StringFlag{
				Name:     "path",
				Usage:    `File to serve`,
				Required: true,
			},
		),
		Action: func(c *cli.Context) error {
			name, path, err := cliNamePath(c)
			if err != nil {
				return err
			}
			o, _, ll := cliToOpts(c)

			opts, err := services.ShareFile(ll, time.Duration(c.Int("ledger-announce-interval"))*time.Second, name, path)
			if err != nil {
				return err
			}
			o = append(o, opts...)

			e, err := node.New(o...)
			if err != nil {
				return err
			}

			displayStart(ll)

			// Start the node to the network, using our ledger
			if err := e.Start(context.Background()); err != nil {
				return err
			}

			for {
				time.Sleep(2 * time.Second)
			}
		},
	}
}

func FileReceive() cli.Command {
	return cli.Command{
		Name:        "file-receive",
		Aliases:     []string{"fr"},
		Usage:       "Receive a file which is served from the network",
		Description: `Receive a file from the network without connecting over VPN`,
		UsageText:   "vpnsvr file-receive unique-id /dst/path",
		Flags: append(CommonFlags,
			cli.StringFlag{
				Name:  "name",
				Usage: `Unique name of the file to be received over the network.`,
			},
			cli.StringFlag{
				Name:  "path",
				Usage: `Destination where to save the file`,
			},
		),
		Action: func(c *cli.Context) error {
			name, path, err := cliNamePath(c)
			if err != nil {
				return err
			}
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

			ledger, _ := e.Ledger()

			return services.ReceiveFile(context.Background(), ledger, e, ll, time.Duration(c.Int("ledger-announce-interval"))*time.Second, name, path)
		},
	}
}
