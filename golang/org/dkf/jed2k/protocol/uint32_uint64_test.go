package protocol

import (
	"bytes"
	"testing"
)

func TestUInt32Basic(t *testing.T) {
	u := NewUInt32(0x12345678)
	
	if u.Uint32Value() != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%x", u.Uint32Value())
	}
	
	if u.BytesCount() != 4 {
		t.Errorf("Expected bytes count 4, got %d", u.BytesCount())
	}
}

func TestUInt32Serialize(t *testing.T) {
	value := NewUInt32(0x12345678)
	buf := &bytes.Buffer{}
	
	err := value.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put UInt32: %v", err)
	}
	
	if buf.Len() != 4 {
		t.Errorf("Expected buffer length 4, got %d", buf.Len())
	}
	
	// Read back
	result := NewUInt32(0)
	err = result.Get(buf)
	if err != nil {
		t.Fatalf("Failed to get UInt32: %v", err)
	}
	
	if result.Uint32Value() != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%x", result.Uint32Value())
	}
}

func TestUInt32Endian(t *testing.T) {
	// Test little endian byte order
	value := NewUInt32(0x12345678)
	buf := &bytes.Buffer{}
	
	err := value.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put UInt32: %v", err)
	}
	
	data := buf.Bytes()
	if len(data) != 4 {
		t.Fatalf("Expected 4 bytes, got %d", len(data))
	}
	
	// Little endian: 0x12345678 should be stored as [0x78, 0x56, 0x34, 0x12]
	expected := []byte{0x78, 0x56, 0x34, 0x12}
	for i, b := range expected {
		if data[i] != b {
			t.Errorf("Byte %d: expected 0x%02x, got 0x%02x", i, b, data[i])
		}
	}
}

func TestUInt64Basic(t *testing.T) {
	u := NewUInt64(0x123456789ABCDEF0)
	
	if u.Uint64Value() != 0x123456789ABCDEF0 {
		t.Errorf("Expected 0x123456789ABCDEF0, got 0x%x", u.Uint64Value())
	}
	
	if u.BytesCount() != 8 {
		t.Errorf("Expected bytes count 8, got %d", u.BytesCount())
	}
}

func TestUInt64Serialize(t *testing.T) {
	value := NewUInt64(0x123456789ABCDEF0)
	buf := &bytes.Buffer{}
	
	err := value.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put UInt64: %v", err)
	}
	
	if buf.Len() != 8 {
		t.Errorf("Expected buffer length 8, got %d", buf.Len())
	}
	
	// Read back
	result := NewUInt64(0)
	err = result.Get(buf)
	if err != nil {
		t.Fatalf("Failed to get UInt64: %v", err)
	}
	
	if result.Uint64Value() != 0x123456789ABCDEF0 {
		t.Errorf("Expected 0x123456789ABCDEF0, got 0x%x", result.Uint64Value())
	}
}