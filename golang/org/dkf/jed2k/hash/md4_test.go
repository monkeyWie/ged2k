package hash

import (
	"testing"
)

func TestMD4Hash(t *testing.T) {
	// Test empty string MD4 hash
	data := []byte("")
	expected := "31d6cfe0d16ae931b73c59d7e0c089c0"
	
	hash := Hash(data)
	if len(hash) != MD4HashSize {
		t.Errorf("Expected hash length %d, got %d", MD4HashSize, len(hash))
	}
	
	// Convert to hex string
	result := ""
	for _, b := range hash {
		result += string("0123456789abcdef"[b>>4])
		result += string("0123456789abcdef"[b&0xf])
	}
	
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestMD4Hasher(t *testing.T) {
	hasher := NewMD4()
	
	// Test basic functionality
	data := []byte("hello world")
	n, err := hasher.Write(data)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	
	hash := hasher.Sum()
	if len(hash) != MD4HashSize {
		t.Errorf("Expected hash length %d, got %d", MD4HashSize, len(hash))
	}
	
	// Test size and block size
	if hasher.Size() != MD4HashSize {
		t.Errorf("Expected size %d, got %d", MD4HashSize, hasher.Size())
	}
	
	if hasher.BlockSize() <= 0 {
		t.Error("Block size should be positive")
	}
}

func TestMD4Reset(t *testing.T) {
	hasher := NewMD4()
	
	// Hash some data
	hasher.Write([]byte("first"))
	hash1 := hasher.Sum()
	
	// Reset and hash different data
	hasher.Reset()
	hasher.Write([]byte("second"))
	hash2 := hasher.Sum()
	
	// They should be different
	equal := true
	for i := 0; i < len(hash1); i++ {
		if hash1[i] != hash2[i] {
			equal = false
			break
		}
	}
	
	if equal {
		t.Error("Hashes should be different after reset")
	}
}