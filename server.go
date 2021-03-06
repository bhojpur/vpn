//go:build !server
// +build !server

package main

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
	"os"

	"github.com/urfave/cli"

	cmd "github.com/bhojpur/vpn/cmd/server"
	internal "github.com/bhojpur/vpn/pkg/version"
)

func main() {
	app := &cli.App{
		Name:        "vpnsvr",
		Version:     internal.Version,
		Author:      "Bhojpur Consulting Private Limited, India",
		Usage:       "vpnsvr --config /etc/bhojpur/vpn/config.yaml",
		Description: "Bhojpur VPN uses libp2p to build an immutable trusted blockchain addressable p2p network",
		Copyright:   cmd.Copyright,
		Flags:       cmd.MainFlags(),
		Commands: []cli.Command{
			cmd.Start(),
			cmd.API(),
			cmd.ServiceAdd(),
			cmd.ServiceConnect(),
			cmd.FileReceive(),
			cmd.Proxy(),
			cmd.FileSend(),
			cmd.DNS(),
			cmd.Peergate(),
		},

		Action: cmd.Main(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
