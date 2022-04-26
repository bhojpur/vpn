package services

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
	"io"
	"net"
	"time"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/node"
	protocol "github.com/bhojpur/vpn/pkg/protocol"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"

	"github.com/bhojpur/vpn/pkg/types"
)

func ExposeNetworkService(announcetime time.Duration, serviceID string) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {
		b.Announce(
			ctx,
			announcetime,
			func() {
				// Retrieve current ID for ip in the blockchain
				existingValue, found := b.GetKey(protocol.ServicesLedgerKey, serviceID)
				service := &types.Service{}
				existingValue.Unmarshal(service)
				// If mismatch, update the blockchain
				if !found || service.PeerID != n.Host().ID().String() {
					updatedMap := map[string]interface{}{}
					updatedMap[serviceID] = types.Service{PeerID: n.Host().ID().String(), Name: serviceID}
					b.Add(protocol.ServicesLedgerKey, updatedMap)
				}
			},
		)
		return nil
	}
}

// ExposeService exposes a service to the p2p network.
// meant to be called before a node is started with Start()
func RegisterService(ll log.StandardLogger, announcetime time.Duration, serviceID, dstaddress string) []node.Option {
	ll.Infof("Exposing service '%s' (%s)", serviceID, dstaddress)
	return []node.Option{
		node.WithStreamHandler(protocol.ServiceProtocol, func(n *node.Node, l *blockchain.Ledger) func(stream network.Stream) {
			return func(stream network.Stream) {
				go func() {
					ll.Infof("(service %s) Received connection from %s", serviceID, stream.Conn().RemotePeer().String())

					// Retrieve current ID for ip in the blockchain
					_, found := l.GetKey(protocol.UsersLedgerKey, stream.Conn().RemotePeer().String())
					// If mismatch, update the blockchain
					if !found {
						ll.Debugf("Reset '%s': not found in the ledger", stream.Conn().RemotePeer().String())
						stream.Reset()
						return
					}

					ll.Infof("Connecting to '%s'", dstaddress)
					c, err := net.Dial("tcp", dstaddress)
					if err != nil {
						ll.Debugf("Reset %s: %s", stream.Conn().RemotePeer().String(), err.Error())
						stream.Reset()
						return
					}
					closer := make(chan struct{}, 2)
					go copyStream(closer, stream, c)
					go copyStream(closer, c, stream)
					<-closer

					stream.Close()
					c.Close()
					ll.Infof("(service %s) Handled correctly '%s'", serviceID, stream.Conn().RemotePeer().String())
				}()
			}
		}),
		node.WithNetworkService(ExposeNetworkService(announcetime, serviceID))}
}

// ConnectNetworkService returns a network service that binds to a service
func ConnectNetworkService(announcetime time.Duration, serviceID string, srcaddr string) node.NetworkService {
	return func(ctx context.Context, c node.Config, node *node.Node, ledger *blockchain.Ledger) error {
		// Open local port for listening
		l, err := net.Listen("tcp", srcaddr)
		if err != nil {
			return err
		}
		//	ll.Info("Binding local port on", srcaddr)

		// Announce ourselves so nodes accepts our connection
		ledger.Announce(
			ctx,
			announcetime,
			func() {
				// Retrieve current ID for ip in the blockchain
				_, found := ledger.GetKey(protocol.UsersLedgerKey, node.Host().ID().String())
				// If mismatch, update the blockchain
				if !found {
					updatedMap := map[string]interface{}{}
					updatedMap[node.Host().ID().String()] = &types.User{
						PeerID:    node.Host().ID().String(),
						Timestamp: time.Now().String(),
					}
					ledger.Add(protocol.UsersLedgerKey, updatedMap)
				}
			},
		)

		defer l.Close()
		for {
			select {
			case <-ctx.Done():
				return errors.New("context canceled")
			default:
				// Listen for an incoming connection.
				conn, err := l.Accept()
				if err != nil {
					//	ll.Error("Error accepting: ", err.Error())
					continue
				}

				//	ll.Info("New connection from", l.Addr().String())
				// Handle connections in a new goroutine, forwarding to the p2p service
				go func() {
					// Retrieve current ID for ip in the blockchain
					existingValue, found := ledger.GetKey(protocol.ServicesLedgerKey, serviceID)
					service := &types.Service{}
					existingValue.Unmarshal(service)
					// If mismatch, update the blockchain
					if !found {
						conn.Close()
						//	ll.Debugf("service '%s' not found on blockchain", serviceID)
						return
					}

					// Decode the Peer
					d, err := peer.Decode(service.PeerID)
					if err != nil {
						conn.Close()
						//	ll.Debugf("could not decode peer '%s'", service.PeerID)
						return
					}

					// Open a stream
					stream, err := node.Host().NewStream(ctx, d, protocol.ServiceProtocol.ID())
					if err != nil {
						conn.Close()
						//	ll.Debugf("could not open stream '%s'", err.Error())
						return
					}
					//	ll.Debugf("(service %s) Redirecting", serviceID, l.Addr().String())

					closer := make(chan struct{}, 2)
					go copyStream(closer, stream, conn)
					go copyStream(closer, conn, stream)
					<-closer

					stream.Close()
					conn.Close()
					//	ll.Infof("(service %s) Done handling %s", serviceID, l.Addr().String())
				}()
			}
		}

	}
}

func copyStream(closer chan struct{}, dst io.Writer, src io.Reader) {
	defer func() { closer <- struct{}{} }() // connection is closed, send signal to stop proxy
	io.Copy(dst, src)
}
