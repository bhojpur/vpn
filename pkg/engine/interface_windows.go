//go:build windows
// +build windows

package engine

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
	"log"
	"net"
	"os/exec"

	"github.com/songgao/water"
)

func prepareInterface(c *Config) error {
	err := netsh("interface", "ip", "set", "address", "name=", c.InterfaceName, "static", c.InterfaceAddress)
	if err != nil {
		log.Println(err)
	}
	err = netsh("interface", "ipv4", "set", "subinterface", c.InterfaceName, "mtu=", fmt.Sprintf("%d", c.InterfaceMTU))
	if err != nil {
		log.Println(err)
	}
	return nil
}

func createInterface(c *Config) (*water.Interface, error) {
	// TUN on Windows requires address and network to be set on device creation stage
	// We also set network to 0.0.0.0/0 so we able to reach networks behind the node
	// https://github.com/songgao/water/blob/master/params_windows.go
	// https://gitlab.com/openconnect/openconnect/-/blob/master/tun-win32.c
	ip, _, err := net.ParseCIDR(c.InterfaceAddress)
	if err != nil {
		return nil, err
	}
	network := net.IPNet{
		IP:   ip,
		Mask: net.IPv4Mask(0, 0, 0, 0),
	}
	config := water.Config{
		DeviceType: c.DeviceType,
		PlatformSpecificParams: water.PlatformSpecificParams{
			ComponentID:   "tap0901",
			InterfaceName: c.InterfaceName,
			Network:       network.String(),
		},
	}

	return water.New(config)
}

func netsh(args ...string) (err error) {
	cmd := exec.Command("netsh", args...)
	err = cmd.Run()
	return
}
