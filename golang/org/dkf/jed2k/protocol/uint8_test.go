package protocol

import (
	"bytes"
	"testing"
)

func TestUInt8Value(t *testing.T) {
	u1 := NewUInt8(0)
	u2 := NewUInt8(0)
	if !u1.Equals(u2) {
		t.Errorf("Expected UInt8(0) to equal UInt8(0)")
	}
}

func TestUInt8Comparable(t *testing.T) {
	tests := []struct {
		a, b     uint8
		expected int
	}{
		{0xff, 0, 1},
		{0xaa, 0xff, -1},
		{250, 250, 0},
	}

	for _, test := range tests {
		u1 := NewUInt8(test.a)
		u2 := NewUInt8(test.b)
		result := u1.CompareTo(u2)
		if result != test.expected {
			t.Errorf("UInt8(%d).CompareTo(UInt8(%d)) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestUInt8Overflow(t *testing.T) {
	u1 := NewUInt8FromInt(0xff0f)
	u2 := NewUInt8(0x0f)
	if u1.CompareTo(u2) != 0 {
		t.Errorf("UInt8 overflow test failed: %d != %d", u1.Uint8Value(), u2.Uint8Value())
	}
}

func TestUInt8Serialize(t *testing.T) {
	data := []byte{1, 2, 3}
	buf := bytes.NewBuffer(data)
	
	value := NewUInt8(0)
	
	// Read first byte
	err := value.Get(buf)
	if err != nil {
		t.Fatalf("Failed to read first byte: %v", err)
	}
	if value.Uint8Value() != 1 {
		t.Errorf("Expected 1, got %d", value.Uint8Value())
	}
	
	// Read second byte
	err = value.Get(buf)
	if err != nil {
		t.Fatalf("Failed to read second byte: %v", err)
	}
	if value.Uint8Value() != 2 {
		t.Errorf("Expected 2, got %d", value.Uint8Value())
	}
	
	// Read third byte
	err = value.Get(buf)
	if err != nil {
		t.Fatalf("Failed to read third byte: %v", err)
	}
	if value.Uint8Value() != 3 {
		t.Errorf("Expected 3, got %d", value.Uint8Value())
	}
	
	// Try to read fourth byte (should fail)
	err = value.Get(buf)
	if err == nil {
		t.Error("Expected error when reading beyond buffer")
	}
}

func TestUInt8Put(t *testing.T) {
	value := NewUInt8(42)
	buf := &bytes.Buffer{}
	
	err := value.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put UInt8: %v", err)
	}
	
	if buf.Len() != 1 {
		t.Errorf("Expected buffer length 1, got %d", buf.Len())
	}
	
	result, err := buf.ReadByte()
	if err != nil {
		t.Fatalf("Failed to read byte from buffer: %v", err)
	}
	
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}