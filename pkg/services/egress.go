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
	"bufio"
	"context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/protocol"
	"github.com/bhojpur/vpn/pkg/types"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

func egressHandler(n *node.Node, b *blockchain.Ledger) func(stream network.Stream) {
	return func(stream network.Stream) {
		// Remember to close the stream when we are done.
		defer stream.Close()

		// Retrieve current ID for ip in the blockchain
		_, found := b.GetKey(protocol.UsersLedgerKey, stream.Conn().RemotePeer().String())
		// If mismatch, update the blockchain
		if !found {
			//		ll.Debugf("Reset '%s': not found in the ledger", stream.Conn().RemotePeer().String())
			stream.Reset()
			return
		}

		// Create a new buffered reader, as ReadRequest needs one.
		// The buffered reader reads from our stream, on which we
		// have sent the HTTP request (see ServeHTTP())
		buf := bufio.NewReader(stream)
		// Read the HTTP request from the buffer
		req, err := http.ReadRequest(buf)
		if err != nil {
			stream.Reset()
			log.Println(err)
			return
		}
		defer req.Body.Close()

		// We need to reset these fields in the request
		// URL as they are not maintained.
		req.URL.Scheme = "http"
		hp := strings.Split(req.Host, ":")
		if len(hp) > 1 && hp[1] == "443" {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}
		req.URL.Host = req.Host

		outreq := new(http.Request)
		*outreq = *req

		// We now make the request
		//fmt.Printf("Making request to %s\n", req.URL)
		resp, err := http.DefaultTransport.RoundTrip(outreq)
		if err != nil {
			stream.Reset()
			log.Println(err)
			return
		}

		// resp.Write writes whatever response we obtained for our
		// request back to the stream.
		resp.Write(stream)
	}
}

// ProxyService starts a local http proxy server which redirects requests to egresses into the network
// It takes a deadtime to consider hosts which are alive within a time window
func ProxyService(announceTime time.Duration, listenAddr string, deadtime time.Duration) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {

		ps := &proxyService{
			host:       n,
			listenAddr: listenAddr,
			deadTime:   deadtime,
		}

		// Announce ourselves so nodes accepts our connection
		b.Announce(
			ctx,
			announceTime,
			func() {
				// Retrieve current ID for ip in the blockchain
				_, found := b.GetKey(protocol.UsersLedgerKey, n.Host().ID().String())
				// If mismatch, update the blockchain
				if !found {
					updatedMap := map[string]interface{}{}
					updatedMap[n.Host().ID().String()] = &types.User{
						PeerID:    n.Host().ID().String(),
						Timestamp: time.Now().String(),
					}
					b.Add(protocol.UsersLedgerKey, updatedMap)
				}
			},
		)

		go ps.Serve()
		return nil
	}
}

type proxyService struct {
	host       *node.Node
	listenAddr string
	deadTime   time.Duration
}

func (p *proxyService) Serve() error {
	return http.ListenAndServe(p.listenAddr, p)
}

func (p *proxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	l, err := p.host.Ledger()
	if err != nil {
		//fmt.Printf("no ledger")
		return
	}

	egress := l.CurrentData()[protocol.EgressService]
	nodes := AvailableNodes(l, p.deadTime)

	availableEgresses := []string{}
	for _, n := range nodes {
		for e := range egress {
			if e == n {
				availableEgresses = append(availableEgresses, e)
			}
		}
	}

	chosen := availableEgresses[rand.Intn(len(availableEgresses)-1)]

	//fmt.Printf("proxying request for %s to peer %s\n", r.URL, chosen)
	// We need to send the request to the remote libp2p peer, so
	// we open a stream to it
	stream, err := p.host.Host().NewStream(context.Background(), peer.ID(chosen), protocol.EgressProtocol.ID())
	// If an error happens, we write an error for response.
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// r.Write() writes the HTTP request to the stream.
	err = r.Write(stream)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Now we read the response that was sent from the dest
	// peer
	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, r)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Copy any headers
	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	// Write response status and headers
	w.WriteHeader(resp.StatusCode)

	// Finally copy the body
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func EgressService(announceTime time.Duration) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {
		b.AnnounceUpdate(ctx, announceTime, protocol.EgressService, n.Host().ID().String(), "ok")
		return nil
	}
}

func Egress(announceTime time.Duration) []node.Option {
	return []node.Option{
		node.WithNetworkService(EgressService(announceTime)),
		node.WithStreamHandler(protocol.EgressProtocol, egressHandler),
	}
}

func Proxy(announceTime, deadtime time.Duration, listenAddr string) []node.Option {
	return []node.Option{
		node.WithNetworkService(ProxyService(announceTime, listenAddr, deadtime)),
	}
}
