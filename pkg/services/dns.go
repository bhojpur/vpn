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
	"fmt"
	"regexp"
	"time"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/node"
	"github.com/bhojpur/vpn/pkg/protocol"
	"github.com/bhojpur/vpn/pkg/types"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-log"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

func DNSNetworkService(ll log.StandardLogger, listenAddr string, forwarder bool, forward []string, cacheSize int) node.NetworkService {
	return func(ctx context.Context, c node.Config, n *node.Node, b *blockchain.Ledger) error {
		server := &dns.Server{Addr: listenAddr, Net: "udp"}
		cache, err := lru.New(cacheSize)
		if err != nil {
			return err
		}
		go func() {
			dns.HandleFunc(".", dnsHandler{ctx, b, forwarder, forward, cache, ll}.handleDNSRequest())
			fmt.Println(server.ListenAndServe())
		}()

		go func() {
			<-ctx.Done()
			server.Shutdown()
		}()

		return nil
	}
}

// DNS returns a network service binding a dns blockchain resolver on listenAddr.
// Takes an associated name for the addresses in the blockchain
func DNS(ll log.StandardLogger, listenAddr string, forwarder bool, forward []string, cacheSize int) []node.Option {
	return []node.Option{
		node.WithNetworkService(DNSNetworkService(ll, listenAddr, forwarder, forward, cacheSize)),
	}
}

// PersistDNSRecord is syntatic sugar around the ledger
// It persists a DNS record to the blockchain until it sees it reconciled.
// It automatically stop announcing and it is not *guaranteed* to persist data.
func PersistDNSRecord(ctx context.Context, b *blockchain.Ledger, announcetime, timeout time.Duration, regex string, record types.DNS) {
	b.Persist(ctx, announcetime, timeout, protocol.DNSKey, regex, record)
}

// AnnounceDNSRecord is syntatic sugar around the ledger
// Announces a DNS record binding to the blockchain, and keeps announcing for the ctx lifecycle
func AnnounceDNSRecord(ctx context.Context, b *blockchain.Ledger, announcetime time.Duration, regex string, record types.DNS) {
	b.AnnounceUpdate(ctx, announcetime, protocol.DNSKey, regex, record)
}

type dnsHandler struct {
	ctx       context.Context
	b         *blockchain.Ledger
	forwarder bool
	forward   []string
	cache     *lru.Cache
	ll        log.StandardLogger
}

func (d dnsHandler) parseQuery(m *dns.Msg, forward bool) *dns.Msg {
	response := m.Copy()
	d.ll.Debug("Received DNS request", m)
	if len(m.Question) > 0 {
		q := m.Question[0]
		// Resolve the entry to an IP from the blockchain data
		for k, v := range d.b.CurrentData()[protocol.DNSKey] {
			r, err := regexp.Compile(k)
			if err == nil && r.MatchString(q.Name) {
				var res types.DNS
				v.Unmarshal(&res)
				if val, exists := res[dns.Type(q.Qtype)]; exists {
					rr, err := dns.NewRR(fmt.Sprintf("%s %s %s", q.Name, dns.TypeToString[q.Qtype], val))
					if err == nil {
						response.Answer = append(m.Answer, rr)
						d.ll.Debug("Response from blockchain", response)
						return response
					}
				}
			}
		}
		if forward {
			d.ll.Debug("Forwarding DNS request", m)
			r, err := d.forwardQuery(m)
			if err == nil {
				response.Answer = r.Answer
			}
			d.ll.Debug("Response from fw server", r)
		}
	}
	return response
}

func (d dnsHandler) handleDNSRequest() func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		var resp *dns.Msg
		switch r.Opcode {
		case dns.OpcodeQuery:
			resp = d.parseQuery(r, d.forwarder)
		}
		resp.SetReply(r)
		resp.Compress = false
		w.WriteMsg(resp)
	}
}

func (d dnsHandler) forwardQuery(dnsMessage *dns.Msg) (*dns.Msg, error) {
	reqCopy := dnsMessage.Copy()
	if len(reqCopy.Question) > 0 {
		if v, ok := d.cache.Get(reqCopy.Question[0].String()); ok {
			q := v.(*dns.Msg)
			q.Id = reqCopy.Id
			return q, nil
		}
	}
	for _, server := range d.forward {
		r, err := QueryDNS(d.ctx, reqCopy, server)
		if r != nil && len(r.Answer) == 0 && !r.MsgHdr.Truncated {
			continue
		}

		if err != nil {
			continue
		}

		if r.Rcode == dns.RcodeSuccess {
			d.cache.Add(reqCopy.Question[0].String(), r)
		}

		if r.Rcode == dns.RcodeNameError || r.Rcode == dns.RcodeSuccess || err == nil {
			return r, err
		}
	}
	return nil, errors.New("not available")
}

// QueryDNS queries a dns server with a dns message and return the answer
// it is blocking.
func QueryDNS(ctx context.Context, msg *dns.Msg, dnsServer string) (*dns.Msg, error) {
	client := &dns.Client{
		Net:            "udp",
		Timeout:        30 * time.Second,
		SingleInflight: true}
	r, _, err := client.Exchange(msg, dnsServer)
	return r, err
}
