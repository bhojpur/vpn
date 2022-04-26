//go:build !windows
// +build !windows

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
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

func createInterface(c *Config) (*water.Interface, error) {
	config := water.Config{
		DeviceType:             c.DeviceType,
		PlatformSpecificParams: water.PlatformSpecificParams{Persist: !c.NetLinkBootstrap},
	}
	config.Name = c.InterfaceName

	return water.New(config)
}

func prepareInterface(c *Config) error {
	link, err := netlink.LinkByName(c.InterfaceName)
	if err != nil {
		return err
	}
	addr, err := netlink.ParseAddr(c.InterfaceAddress)
	if err != nil {
		return err
	}

	err = netlink.LinkSetMTU(link, c.InterfaceMTU)
	if err != nil {
		return err
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		return err
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}
	return nil
}
