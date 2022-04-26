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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type DataString string

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	Storage   map[string]map[string]Data
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
type Blockchain []Block

// make sure block is valid by checking index, and comparing the hash of the previous block
func (newBlock Block) IsValid(oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if newBlock.Checksum() != newBlock.Hash {
		return false
	}

	return true
}

// Checksum does SHA256 hashing of the block
func (b Block) Checksum() string {
	record := fmt.Sprint(b.Index, b.Timestamp, b.Storage, b.PrevHash)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func (oldBlock Block) NewBlock(s map[string]map[string]Data) Block {
	var newBlock Block

	t := time.Now().UTC()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Storage = s
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = newBlock.Checksum()

	return newBlock
}
