package blockchain

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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/peterbourgon/diskv"
)

type DiskStore struct {
	chain *diskv.Diskv
}

func NewDiskStore(d *diskv.Diskv) *DiskStore {
	return &DiskStore{chain: d}
}

func (m *DiskStore) Add(b Block) {
	bb, _ := json.Marshal(b)
	m.chain.Write(fmt.Sprint(b.Index), bb)
	m.chain.Write("index", []byte(fmt.Sprint(b.Index)))

}

func (m *DiskStore) Len() int {
	count, err := m.chain.Read("index")
	if err != nil {
		return 0
	}
	c, _ := strconv.Atoi(string(count))
	return c

}

func (m *DiskStore) Last() Block {
	b := &Block{}

	count, err := m.chain.Read("index")
	if err != nil {
		return *b
	}

	dat, _ := m.chain.Read(string(count))
	json.Unmarshal(dat, b)

	return *b
}
