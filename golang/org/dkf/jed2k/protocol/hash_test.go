package protocol

import (
	"bytes"
	"testing"
)

func TestHashFromString(t *testing.T) {
	// Test creating hash from valid hex string
	hashStr := "31D6CFE0D16AE931B73C59D7E0C089C0"
	hash, err := HashFromString(hashStr)
	if err != nil {
		t.Fatalf("Failed to create hash from string: %v", err)
	}
	
	if hash.String() != hashStr {
		t.Errorf("Expected %s, got %s", hashStr, hash.String())
	}
}

func TestHashFromStringInvalid(t *testing.T) {
	// Test invalid length
	_, err := HashFromString("invalid")
	if err == nil {
		t.Error("Expected error for invalid hash string")
	}
	
	// Test invalid hex characters
	_, err = HashFromString("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ")
	if err == nil {
		t.Error("Expected error for invalid hex characters")
	}
}

func TestHashFromBytes(t *testing.T) {
	data := []byte{0x31, 0xd6, 0xcf, 0xe0, 0xd1, 0x6a, 0xe9, 0x31,
		0xb7, 0x3c, 0x59, 0xd7, 0xe0, 0xc0, 0x89, 0xc0}
	
	hash, err := HashFromBytes(data)
	if err != nil {
		t.Fatalf("Failed to create hash from bytes: %v", err)
	}
	
	expected := "31D6CFE0D16AE931B73C59D7E0C089C0"
	if hash.String() != expected {
		t.Errorf("Expected %s, got %s", expected, hash.String())
	}
}

func TestHashFromBytesInvalid(t *testing.T) {
	// Test invalid length
	data := []byte{1, 2, 3}
	_, err := HashFromBytes(data)
	if err == nil {
		t.Error("Expected error for invalid byte length")
	}
}

func TestHashSerialize(t *testing.T) {
	// Create a hash
	original := MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	
	// Serialize to buffer
	buf := &bytes.Buffer{}
	err := original.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put hash: %v", err)
	}
	
	if buf.Len() != HashSize {
		t.Errorf("Expected buffer length %d, got %d", HashSize, buf.Len())
	}
	
	// Deserialize from buffer
	result := NewHash()
	err = result.Get(buf)
	if err != nil {
		t.Fatalf("Failed to get hash: %v", err)
	}
	
	if !original.Equals(result) {
		t.Errorf("Serialization/deserialization failed: %s != %s", original.String(), result.String())
	}
}

func TestHashEquals(t *testing.T) {
	hash1 := MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	hash2 := MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	hash3 := MustHashFromString("31D6CFE0D14CE931B73C59D7E0C04BC0")
	
	if !hash1.Equals(hash2) {
		t.Error("Equal hashes should be equal")
	}
	
	if hash1.Equals(hash3) {
		t.Error("Different hashes should not be equal")
	}
}

func TestHashCompareTo(t *testing.T) {
	hash1 := MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	hash2 := MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	hash3 := MustHashFromString("41D6CFE0D16AE931B73C59D7E0C089C0")
	
	if hash1.CompareTo(hash2) != 0 {
		t.Error("Equal hashes should compare to 0")
	}
	
	if hash1.CompareTo(hash3) >= 0 {
		t.Error("hash1 should be less than hash3")
	}
	
	if hash3.CompareTo(hash1) <= 0 {
		t.Error("hash3 should be greater than hash1")
	}
}

func TestHashIsZero(t *testing.T) {
	hash := NewHash()
	if !hash.IsZero() {
		t.Error("New hash should be zero")
	}
	
	hash = MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	if hash.IsZero() {
		t.Error("Non-zero hash should not be zero")
	}
}

func TestHashClear(t *testing.T) {
	hash := MustHashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
	if hash.IsZero() {
		t.Error("Hash should not be zero initially")
	}
	
	hash.Clear()
	if !hash.IsZero() {
		t.Error("Hash should be zero after clear")
	}
}

func TestPredefinedHashes(t *testing.T) {
	// Test that predefined hashes are valid
	if TerminalHash.IsZero() {
		t.Error("TerminalHash should not be zero")
	}
	
	if LibEd2kHash.IsZero() {
		t.Error("LibEd2kHash should not be zero")
	}
	
	if EmuleHash.IsZero() {
		t.Error("EmuleHash should not be zero")
	}
	
	if !InvalidHash.IsZero() {
		t.Error("InvalidHash should be zero")
	}
}