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
	"os"
	"time"

	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/protocol"
	"github.com/ipfs/go-log"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/types"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

func SharefileNetworkService(announcetime time.Duration, fileID string) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {
		// By announcing periodically our service to the blockchain
		b.Announce(
			ctx,
			announcetime,
			func() {
				// Retrieve current ID for ip in the blockchain
				existingValue, found := b.GetKey(protocol.FilesLedgerKey, fileID)
				service := &types.Service{}
				existingValue.Unmarshal(service)
				// If mismatch, update the blockchain
				if !found || service.PeerID != n.Host().ID().String() {
					updatedMap := map[string]interface{}{}
					updatedMap[fileID] = types.File{PeerID: n.Host().ID().String(), Name: fileID}
					b.Add(protocol.FilesLedgerKey, updatedMap)
				}
			},
		)
		return nil
	}
}

// ShareFile shares a file to the p2p network.
// meant to be called before a node is started with Start()
func ShareFile(ll log.StandardLogger, announcetime time.Duration, fileID, filepath string) ([]node.Option, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	ll.Infof("Serving '%s' as '%s'", filepath, fileID)
	return []node.Option{
		node.WithNetworkService(
			SharefileNetworkService(announcetime, fileID),
		),
		node.WithStreamHandler(protocol.FileProtocol,
			func(n *node.Node, l *blockchain.Ledger) func(stream network.Stream) {
				return func(stream network.Stream) {
					go func() {
						ll.Infof("(file %s) Received connection from %s", fileID, stream.Conn().RemotePeer().String())

						// Retrieve current ID for ip in the blockchain
						_, found := l.GetKey(protocol.UsersLedgerKey, stream.Conn().RemotePeer().String())
						// If mismatch, update the blockchain
						if !found {
							ll.Info("Reset", stream.Conn().RemotePeer().String(), "Not found in the ledger")
							stream.Reset()
							return
						}
						f, err := os.Open(filepath)
						if err != nil {
							return
						}
						io.Copy(stream, f)
						f.Close()
						stream.Close()

						ll.Infof("(file %s) Done handling %s", fileID, stream.Conn().RemotePeer().String())
					}()
				}
			})}, nil

}

func ReceiveFile(ctx context.Context, ledger *blockchain.Ledger, n *node.Node, l log.StandardLogger, announcetime time.Duration, fileID string, path string) error {
	// Announce ourselves so nodes accepts our connection
	ledger.Announce(
		ctx,
		announcetime,
		func() {
			// Retrieve current ID for ip in the blockchain
			_, found := ledger.GetKey(protocol.UsersLedgerKey, n.Host().ID().String())

			// If mismatch, update the blockchain
			if !found {
				updatedMap := map[string]interface{}{}
				updatedMap[n.Host().ID().String()] = &types.User{
					PeerID:    n.Host().ID().String(),
					Timestamp: time.Now().String(),
				}
				ledger.Add(protocol.UsersLedgerKey, updatedMap)
			}
		},
	)

	for {
		select {
		case <-ctx.Done():
			return errors.New("context canceled")
		default:
			time.Sleep(5 * time.Second)

			l.Debug("Attempting to find file in the blockchain")

			existingValue, found := ledger.GetKey(protocol.FilesLedgerKey, fileID)
			fi := &types.File{}
			existingValue.Unmarshal(fi)
			// If mismatch, update the blockchain
			if !found {
				l.Debug("file not found on blockchain, retrying in 5 seconds")
				continue
			} else {
				// Retrieve current ID for ip in the blockchain
				existingValue, found := ledger.GetKey(protocol.FilesLedgerKey, fileID)
				fi := &types.File{}
				existingValue.Unmarshal(fi)

				// If mismatch, update the blockchain
				if !found {
					return errors.New("file not found")
				}

				// Decode the Peer
				d, err := peer.Decode(fi.PeerID)
				if err != nil {
					return err
				}

				l.Debug("file found on blockchain, opening stream to", d)

				// Open a stream
				stream, err := n.Host().NewStream(ctx, d, protocol.FileProtocol.ID())
				if err != nil {
					l.Debugf("failed to dial %s, retrying in 5 seconds", d)
					continue
				}

				l.Infof("Saving file %s to %s", fileID, path)

				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					return err
				}

				io.Copy(f, stream)
				f.Close()

				l.Infof("Received file %s to %s", fileID, path)
				return nil
			}
		}
	}
}
