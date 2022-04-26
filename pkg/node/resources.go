package node

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
	"strings"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pprotocol "github.com/libp2p/go-libp2p-core/protocol"
	rcmgr "github.com/libp2p/go-libp2p-resource-manager"
)

type NetLimitConfig struct {
	Dynamic bool `json:",omitempty"`
	// set if Dynamic is false
	Memory int64 `json:",omitempty"`
	// set if Dynamic is true
	MemoryFraction float64 `json:",omitempty"`
	MinMemory      int64   `json:",omitempty"`
	MaxMemory      int64   `json:",omitempty"`

	Streams, StreamsInbound, StreamsOutbound int
	Conns, ConnsInbound, ConnsOutbound       int
	FD                                       int
}

func NetSetLimit(mgr network.ResourceManager, scope string, limit NetLimitConfig) error {
	setLimit := func(s network.ResourceScope) error {
		limiter, ok := s.(rcmgr.ResourceScopeLimiter)
		if !ok {
			return fmt.Errorf("resource scope doesn't implement ResourceScopeLimiter interface")
		}

		var newLimit rcmgr.Limit
		if limit.Dynamic {
			newLimit = &rcmgr.DynamicLimit{
				MemoryLimit: rcmgr.MemoryLimit{
					MemoryFraction: limit.MemoryFraction,
					MinMemory:      limit.MinMemory,
					MaxMemory:      limit.MaxMemory,
				},
				BaseLimit: rcmgr.BaseLimit{
					Streams:         limit.Streams,
					StreamsInbound:  limit.StreamsInbound,
					StreamsOutbound: limit.StreamsOutbound,
					Conns:           limit.Conns,
					ConnsInbound:    limit.ConnsInbound,
					ConnsOutbound:   limit.ConnsOutbound,
					FD:              limit.FD,
				},
			}
		} else {
			newLimit = &rcmgr.StaticLimit{
				Memory: limit.Memory,
				BaseLimit: rcmgr.BaseLimit{
					Streams:         limit.Streams,
					StreamsInbound:  limit.StreamsInbound,
					StreamsOutbound: limit.StreamsOutbound,
					Conns:           limit.Conns,
					ConnsInbound:    limit.ConnsInbound,
					ConnsOutbound:   limit.ConnsOutbound,
					FD:              limit.FD,
				},
			}
		}

		limiter.SetLimit(newLimit)
		return nil
	}

	switch {
	case scope == "system":
		err := mgr.ViewSystem(func(s network.ResourceScope) error {
			return setLimit(s)
		})
		return err

	case scope == "transient":
		err := mgr.ViewTransient(func(s network.ResourceScope) error {
			return setLimit(s)
		})
		return err

	case strings.HasPrefix(scope, "svc:"):
		svc := scope[4:]
		err := mgr.ViewService(svc, func(s network.ServiceScope) error {
			return setLimit(s)
		})
		return err

	case strings.HasPrefix(scope, "proto:"):
		proto := scope[6:]
		err := mgr.ViewProtocol(libp2pprotocol.ID(proto), func(s network.ProtocolScope) error {
			return setLimit(s)
		})
		return err

	case strings.HasPrefix(scope, "peer:"):
		p := scope[5:]
		pid, err := peer.Decode(p)
		if err != nil {
			return fmt.Errorf("invalid peer ID: %s: %w", p, err)
		}
		err = mgr.ViewPeer(pid, func(s network.PeerScope) error {
			return setLimit(s)
		})
		return err

	default:
		return fmt.Errorf("invalid scope %s", scope)
	}
}
