package services_test

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
	"time"

	"github.com/ipfs/go-log"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/logger"
	node "github.com/bhojpur/vpn/pkg/node"
	. "github.com/bhojpur/vpn/pkg/services"
	"github.com/bhojpur/vpn/pkg/types"
)

var _ = Describe("DNS service", func() {
	token := node.GenerateNewConnectionData().Base64()

	logg := logger.New(log.LevelDebug)
	l := node.Logger(logg)

	e2, _ := node.New(
		append(Alive(15*time.Second, 90*time.Minute, 15*time.Minute),
			node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)...)

	Context("DNS service", func() {
		It("Set DNS records and can resolve IPs", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			opts := DNS(logg, "127.0.0.1:19192", true, []string{"8.8.8.8:53"}, 10)
			opts = append(opts, node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)
			e, _ := node.New(opts...)

			e.Start(ctx)
			e2.Start(ctx)

			ll, _ := e2.Ledger()

			AnnounceDNSRecord(ctx, ll, 60*time.Second, `test.foo.`, types.DNS{
				dns.Type(dns.TypeA): "2.2.2.2",
			})

			searchDomain := func(d string) func() string {
				return func() string {
					var s string
					dnsMessage := new(dns.Msg)
					dnsMessage.SetQuestion(fmt.Sprintf("%s.", d), dns.TypeA)

					r, err := QueryDNS(ctx, dnsMessage, "127.0.0.1:19192")
					if r != nil {
						answers := r.Answer
						for _, a := range answers {

							s = a.String() + s
						}
					}
					if err != nil {
						fmt.Println(err)
					}
					return s
				}
			}

			Eventually(searchDomain("google.com"), 230*time.Second, 1*time.Second).Should(ContainSubstring("A"))
			// We hit the same record again, this time it's faster as there is a cache
			Eventually(searchDomain("google.com"), 1*time.Second, 1*time.Second).Should(ContainSubstring("A"))
			Eventually(searchDomain("test.foo"), 230*time.Second, 1*time.Second).Should(ContainSubstring("2.2.2.2"))
		})
	})
})
