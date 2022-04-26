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

import "encoding/json"

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message  string
	SenderID string

	Annotations map[string]interface{}
}

type MessageOption func(cfg *Message) error

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (m *Message) Apply(opts ...MessageOption) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(m); err != nil {
			return err
		}
	}
	return nil
}

func NewMessage(s string) *Message {
	return &Message{Message: s}
}

func (m *Message) Copy() *Message {
	copy := *m
	return &copy
}

func (m *Message) WithMessage(s string) *Message {
	copy := m.Copy()
	copy.Message = s
	return copy
}

func (m *Message) AnnotationsToObj(v interface{}) error {
	blob, err := json.Marshal(m.Annotations)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, v)
}
