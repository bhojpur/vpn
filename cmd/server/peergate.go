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
	"fmt"

	"github.com/bhojpur/vpn/pkg/trustzone/authprovider/ecdsa"
	"github.com/urfave/cli"
)

func Peergate() cli.Command {
	return cli.Command{
		Name:        "peergater",
		Usage:       "peergater ecdsa-genkey",
		Description: `Peergater auth utilities`,
		Subcommands: cli.Commands{
			{
				Name: "ecdsa-genkey",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "privkey",
					},
					&cli.BoolFlag{
						Name: "pubkey",
					},
				},
				Action: func(c *cli.Context) error {
					priv, pub, err := ecdsa.GenerateKeys()
					if !c.Bool("privkey") && !c.Bool("pubkey") {
						fmt.Printf("Private key: %s\n", string(priv))
						fmt.Printf("Public key: %s\n", string(pub))
					} else if c.Bool("privkey") {
						fmt.Printf(string(priv))
					} else if c.Bool("pubkey") {
						fmt.Printf(string(pub))
					}
					return err
				},
			},
		},
	}
}
