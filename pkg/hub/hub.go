package hub

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
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/bhojpur/vpn/pkg/crypto"
	"github.com/xlzd/gotp"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type MessageHub struct {
	sync.Mutex

	blockchain, public *room
	ps                 *pubsub.PubSub
	otpKey             string
	maxsize            int
	keyLength          int
	interval           int
	joinPublic         bool

	ctxCancel                context.CancelFunc
	Messages, PublicMessages chan *Message
}

// roomBufSize is the number of incoming messages to buffer for each topic.
const roomBufSize = 128

func NewHub(otp string, maxsize, keyLength, interval int, joinPublic bool) *MessageHub {
	return &MessageHub{otpKey: otp, maxsize: maxsize, keyLength: keyLength, interval: interval,
		Messages: make(chan *Message, roomBufSize), PublicMessages: make(chan *Message, roomBufSize), joinPublic: joinPublic}
}

func (m *MessageHub) topicKey(salts ...string) string {
	totp := gotp.NewTOTP(strings.ToUpper(m.otpKey), m.keyLength, m.interval, nil)
	if len(salts) > 0 {
		return crypto.MD5(totp.Now() + strings.Join(salts, ":"))
	}
	return crypto.MD5(totp.Now())
}

func (m *MessageHub) joinRoom(host host.Host) error {
	m.Lock()
	defer m.Unlock()

	if m.ctxCancel != nil {
		m.ctxCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.ctxCancel = cancel

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, host, pubsub.WithMaxMessageSize(m.maxsize))
	if err != nil {
		return err
	}

	// join the "chat" room
	cr, err := connect(ctx, ps, host.ID(), m.topicKey(), m.Messages)
	if err != nil {
		return err
	}

	m.blockchain = cr

	if m.joinPublic {
		cr2, err := connect(ctx, ps, host.ID(), m.topicKey("public"), m.PublicMessages)
		if err != nil {
			return err
		}
		m.public = cr2
	}

	m.ps = ps

	return nil
}

func (m *MessageHub) Start(ctx context.Context, host host.Host) error {
	c := make(chan interface{})
	go func(c context.Context, cc chan interface{}) {
		k := ""
		for {
			select {
			default:
				currentKey := m.topicKey()
				if currentKey != k {
					k = currentKey
					cc <- nil
				}
				time.Sleep(1 * time.Second)
			case <-ctx.Done():
				close(cc)
				return
			}
		}
	}(ctx, c)

	for range c {
		m.joinRoom(host)
	}

	// Close eventual open contexts
	if m.ctxCancel != nil {
		m.ctxCancel()
	}
	return nil
}

func (m *MessageHub) PublishMessage(mess *Message) error {
	m.Lock()
	defer m.Unlock()
	if m.blockchain != nil {
		return m.blockchain.publishMessage(mess)
	}
	return errors.New("no message room available")
}

func (m *MessageHub) PublishPublicMessage(mess *Message) error {
	m.Lock()
	defer m.Unlock()
	if m.public != nil {
		return m.public.publishMessage(mess)
	}
	return errors.New("no message room available")
}

func (m *MessageHub) ListPeers() ([]peer.ID, error) {
	m.Lock()
	defer m.Unlock()
	if m.blockchain != nil {
		return m.blockchain.Topic.ListPeers(), nil
	}
	return nil, errors.New("no message room available")
}
