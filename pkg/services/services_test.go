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
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ipfs/go-log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/logger"
	node "github.com/bhojpur/vpn/pkg/node"
	. "github.com/bhojpur/vpn/pkg/services"
)

func get(url string) string {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

var _ = Describe("Expose services", func() {
	token := node.GenerateNewConnectionData().Base64()

	logg := logger.New(log.LevelFatal)
	l := node.Logger(logg)
	serviceUUID := "test"

	Context("Service sharing", func() {
		It("expose services and can connect to them", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			opts := RegisterService(logg, 5*time.Second, serviceUUID, "142.250.184.35:80")
			opts = append(opts, node.FromBase64(true, true, token), node.WithDiscoveryInterval(10*time.Second), node.WithStore(&blockchain.MemoryStore{}), l)
			e, _ := node.New(opts...)

			// First node expose a service
			// redirects to google:80

			e.Start(ctx)

			go func() {
				e2, _ := node.New(
					node.WithNetworkService(ConnectNetworkService(5*time.Second, serviceUUID, "127.0.0.1:9999")),
					node.WithDiscoveryInterval(10*time.Second),
					node.FromBase64(true, true, token), node.WithStore(&blockchain.MemoryStore{}), l)

				e2.Start(ctx)
			}()

			Eventually(func() string {
				return get("http://127.0.0.1:9999")
			}, 360*time.Second, 1*time.Second).Should(ContainSubstring("The document has moved"))
		})
	})
})
