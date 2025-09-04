package protocol

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const HashSize = 16 // MD4 hash size in bytes

// Hash represents a 16-byte MD4 hash used in ed2k protocol
type Hash struct {
	value [HashSize]byte
}

// Predefined hashes
var (
	TerminalHash = MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	LibEd2kHash  = MustHashFromString("31D6CFE0D14CE931B73C59D7E0C04BC0")
	EmuleHash    = MustHashFromString("31D6CFE0D10EE931B73C59D7E0C06FC0")
	InvalidHash  = &Hash{}
)

// NewHash creates a new empty hash
func NewHash() *Hash {
	return &Hash{}
}

// NewHashFromHash creates a copy of another hash
func NewHashFromHash(h *Hash) *Hash {
	result := &Hash{}
	copy(result.value[:], h.value[:])
	return result
}

// Assign copies the value from another hash
func (h *Hash) Assign(other *Hash) *Hash {
	copy(h.value[:], other.value[:])
	return h
}

// HashFromString creates a hash from a hex string
func HashFromString(value string) (*Hash, error) {
	if len(value) != HashSize*2 {
		return InvalidHash, fmt.Errorf("invalid hash string length: expected %d, got %d", HashSize*2, len(value))
	}
	
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return InvalidHash, err
	}
	
	hash := &Hash{}
	copy(hash.value[:], bytes)
	return hash, nil
}

// MustHashFromString creates a hash from hex string, panics on error
func MustHashFromString(value string) *Hash {
	hash, err := HashFromString(value)
	if err != nil {
		panic(err)
	}
	return hash
}

// HashFromBytes creates a hash from byte slice
func HashFromBytes(value []byte) (*Hash, error) {
	if len(value) != HashSize {
		return InvalidHash, fmt.Errorf("invalid hash byte length: expected %d, got %d", HashSize, len(value))
	}
	
	hash := &Hash{}
	copy(hash.value[:], value)
	return hash, nil
}

// Get deserializes from buffer
func (h *Hash) Get(src *bytes.Buffer) error {
	if src.Len() < HashSize {
		return fmt.Errorf("buffer underflow: need %d bytes, have %d", HashSize, src.Len())
	}
	
	_, err := src.Read(h.value[:])
	return err
}

// Put serializes to buffer
func (h *Hash) Put(dst *bytes.Buffer) error {
	_, err := dst.Write(h.value[:])
	return err
}

// BytesCount returns the number of bytes this type uses
func (h *Hash) BytesCount() int {
	return HashSize
}

// Bytes returns a copy of the hash bytes
func (h *Hash) Bytes() []byte {
	result := make([]byte, HashSize)
	copy(result, h.value[:])
	return result
}

// String returns hex string representation
func (h *Hash) String() string {
	return strings.ToUpper(hex.EncodeToString(h.value[:]))
}

// Equals checks equality with another hash
func (h *Hash) Equals(other *Hash) bool {
	return bytes.Equal(h.value[:], other.value[:])
}

// CompareTo compares with another hash
func (h *Hash) CompareTo(other *Hash) int {
	return bytes.Compare(h.value[:], other.value[:])
}

// IsZero checks if hash is all zeros
func (h *Hash) IsZero() bool {
	for _, b := range h.value {
		if b != 0 {
			return false
		}
	}
	return true
}

// Clear sets all bytes to zero
func (h *Hash) Clear() {
	for i := range h.value {
		h.value[i] = 0
	}
}