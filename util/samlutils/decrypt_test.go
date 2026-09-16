// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package samlutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"
)

// A ciphertext that is not an IV followed by whole blocks cannot be
// decrypted and must be reported rather than indexed into.
func TestDecryptAesCbcRejectsBadLength(t *testing.T) {
	key := make([]byte, 16)
	cases := []struct {
		name   string
		secret []byte
	}{
		{"empty", []byte{}},
		{"shorter than one block", []byte("AAAA")},
		{"exactly one block", make([]byte, aes.BlockSize)},
		{"one block plus one byte", make([]byte, aes.BlockSize+1)},
		{"not a whole number of blocks", make([]byte, 2*aes.BlockSize+4)},
	}
	for _, c := range cases {
		if _, err := decryptAesCbc(key, c.secret); err == nil {
			t.Errorf("%s: decryptAesCbc returned no error for %d bytes", c.name, len(c.secret))
		}
	}
}

// A well formed payload decrypts and has its padding removed.
func TestDecryptAesCbcRoundTrip(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := []byte(`<saml:Assertion>hello</saml:Assertion>`)

	// Encrypt the way XML Encryption does: IV, then the padded plaintext.
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}
	padLen := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := append(append([]byte(nil), plaintext...), bytes.Repeat([]byte{byte(padLen)}, padLen)...)
	iv := bytes.Repeat([]byte{0x42}, aes.BlockSize)
	secret := append(append([]byte(nil), iv...), make([]byte, len(padded))...)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(secret[aes.BlockSize:], padded)

	got, err := decryptAesCbc(key, secret)
	if err != nil {
		t.Fatalf("decryptAesCbc: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("decryptAesCbc = %q, want %q", got, plaintext)
	}
}

// padded builds a payload of two blocks ending in a run of n bytes of value n.
func padded(blockSize, n int) []byte {
	in := bytes.Repeat([]byte("x"), 2*blockSize)
	for i := 2*blockSize - n; i < 2*blockSize; i++ {
		in[i] = byte(n)
	}
	return in
}

func TestStripPKCS7Padding(t *testing.T) {
	const bs = aes.BlockSize
	for n := 1; n <= bs; n++ {
		in := padded(bs, n)
		got := stripPKCS7Padding(in, bs)
		if len(got) != 2*bs-n {
			t.Errorf("padding %d: length %d, want %d", n, len(got), 2*bs-n)
		}
	}
}

// Data that does not carry valid padding is returned unchanged, including
// input that is not a whole number of blocks.
func TestStripPKCS7PaddingLeavesInvalidAlone(t *testing.T) {
	const bs = aes.BlockSize

	zeroLen := padded(bs, 1)
	zeroLen[len(zeroLen)-1] = 0x00

	tooLong := padded(bs, 1)
	tooLong[len(tooLong)-1] = byte(bs + 1)

	mismatched := padded(bs, 3)
	mismatched[2*bs-1] = 0x02

	unchanged := [][]byte{
		{},
		bytes.Repeat([]byte("x"), bs), // no padding run at all
		zeroLen,                       // padding length 0
		tooLong,                       // padding length beyond the block
		mismatched,                    // run does not agree with its length
		padded(bs, 1)[:2*bs-1],        // not a whole number of blocks
	}
	for i, in := range unchanged {
		if got := stripPKCS7Padding(in, bs); !bytes.Equal(got, in) {
			t.Errorf("case %d: input of %d bytes changed to %d bytes", i, len(in), len(got))
		}
	}
}

// EncryptedData whose KeyInfo carries no EncryptedKey is reported rather
// than dereferenced.
func TestDecryptDataRejectsMissingEncryptedKey(t *testing.T) {
	data := EncryptedData{
		CipherData: CipherData{
			CipherValue: CipherValue{Value: "AAAA"},
		},
		EncryptionMethod: EncryptionMethod{
			Algorithm: "http://www.w3.org/2001/04/xmlenc#aes128-cbc",
		},
	}
	if _, err := data.decryptData(nil); err == nil {
		t.Error("no KeyInfo.EncryptedKey: expected an error")
	}
}

// An OAEP EncryptedKey without a DigestMethod is reported rather than
// dereferenced.
func TestDecryptKeyRejectsMissingDigestMethod(t *testing.T) {
	key := EncryptedKey{
		CipherData: CipherData{CipherValue: CipherValue{Value: "AAAA"}},
		EncryptionMethod: EncryptionMethod{
			Algorithm: "http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p",
		},
	}
	if _, err := key.decryptKey(nil); err == nil {
		t.Error("OAEP without a DigestMethod: expected an error")
	}
}
