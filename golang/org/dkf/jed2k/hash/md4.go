package hash

import (
	"encoding/hex"
	"fmt"
	"hash"
	"golang.org/x/crypto/md4"
)

const MD4HashSize = 16

// Hash represents a 16-byte MD4 hash
type Hash struct {
	bytes [MD4HashSize]byte
}

// NewHash creates a new empty hash
func NewHash() *Hash {
	return &Hash{}
}

// NewHashFromBytes creates a hash from byte array
func NewHashFromBytes(bytes [MD4HashSize]byte) *Hash {
	return &Hash{bytes: bytes}
}

// NewEmuleHash creates a hash with eMule's default values
func NewEmuleHash() *Hash {
	// eMule default hash
	return &Hash{bytes: [MD4HashSize]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x6A, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x89, 0xC0}}
}

// FromHexString creates a hash from hex string
func FromHexString(hexStr string) (*Hash, error) {
	if len(hexStr) != 32 {
		return nil, fmt.Errorf("hex string must be 32 characters long")
	}
	
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %v", err)
	}
	
	if len(bytes) != MD4HashSize {
		return nil, fmt.Errorf("decoded bytes must be %d bytes long", MD4HashSize)
	}
	
	var hashBytes [MD4HashSize]byte
	copy(hashBytes[:], bytes)
	
	return &Hash{bytes: hashBytes}, nil
}

// String returns the hex string representation
func (h *Hash) String() string {
	return hex.EncodeToString(h.bytes[:])
}

// Bytes returns the hash bytes
func (h *Hash) Bytes() [MD4HashSize]byte {
	return h.bytes
}

// Assign sets the hash from another hash
func (h *Hash) Assign(other *Hash) {
	h.bytes = other.bytes
}

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

// ComputeHash computes MD4 hash of data
func ComputeHash(data []byte) []byte {
	h := md4.New()
	h.Write(data)
	return h.Sum(nil)
}