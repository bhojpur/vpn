package crypto_test

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
	. "github.com/bhojpur/vpn/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/bhojpur/vpn/pkg/crypto"
)

var _ = Describe("Crypto utilities", func() {
	Context("AES", func() {
		It("Encode/decode", func() {
			key := RandStringRunes(32)
			message := "foo"
			k := [32]byte{}
			copy([]byte(key)[:], k[:32])

			encoded, err := AESEncrypt(message, &k)
			Expect(err).ToNot(HaveOccurred())
			Expect(encoded).ToNot(Equal(key))
			Expect(len(encoded)).To(Equal(62))

			// Encode again
			encoded2, err := AESEncrypt(message, &k)
			Expect(err).ToNot(HaveOccurred())

			// should differ
			Expect(encoded2).ToNot(Equal(encoded))

			// Decrypt and check
			decoded, err := AESDecrypt(encoded, &k)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(message))

			decoded, err = AESDecrypt(encoded2, &k)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(message))
		})
	})
})
