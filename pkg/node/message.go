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
	hub "github.com/bhojpur/vpn/pkg/hub"
)

// messageWriter is a struct returned by the node that satisfies the io.Writer interface
// on the underlying hub.
// Everything Write into the message writer is enqueued to a message channel
// which is sealed and processed by the node
type messageWriter struct {
	input chan<- *hub.Message
	c     Config
	mess  *hub.Message
}

// Write writes a slice of bytes to the message channel
func (mw *messageWriter) Write(p []byte) (n int, err error) {
	return mw.Send(mw.mess.WithMessage(string(p)))
}

// Send sends a message to the channel
func (mw *messageWriter) Send(copy *hub.Message) (n int, err error) {
	mw.input <- copy
	return len(copy.Message), nil
}
