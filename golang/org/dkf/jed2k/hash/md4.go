package hash

import (
	"hash"
	"golang.org/x/crypto/md4"
)

const MD4HashSize = 16

// MD4 provides MD4 hashing functionality using Go's crypto library
type MD4 struct {
	hasher hash.Hash
}

// NewMD4 creates a new MD4 hasher
func NewMD4() *MD4 {
	return &MD4{hasher: md4.New()}
}

// Write adds data to the hash
func (m *MD4) Write(data []byte) (int, error) {
	return m.hasher.Write(data)
}

// Sum returns the MD4 hash
func (m *MD4) Sum() []byte {
	return m.hasher.Sum(nil)
}

// Reset resets the hasher state
func (m *MD4) Reset() {
	m.hasher.Reset()
}

// Size returns the hash size in bytes
func (m *MD4) Size() int {
	return MD4HashSize
}

// BlockSize returns the block size
func (m *MD4) BlockSize() int {
	return m.hasher.BlockSize()
}

// Hash computes MD4 hash of data
func Hash(data []byte) []byte {
	h := md4.New()
	h.Write(data)
	return h.Sum(nil)
}