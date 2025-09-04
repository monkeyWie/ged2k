package protocol

import (
	"bytes"
	"testing"
)

func TestUInt16Initial(t *testing.T) {
	tests := []struct {
		input    int
		expected uint16
	}{
		{10, 10},
		{0x10, 0x10},
		{0xffff, 0xffff},
	}

	for _, test := range tests {
		u1 := NewUInt16FromInt(test.input)
		u2 := NewUInt16(test.expected)
		if !u1.Equals(u2) {
			t.Errorf("UInt16(%d) != UInt16(%d)", test.input, test.expected)
		}
	}

	// Test default value
	value := NewUInt16(0)
	if value.Uint16Value() != 0 {
		t.Errorf("Expected default value 0, got %d", value.Uint16Value())
	}

	// Test assignment
	value.Assign(1000)
	if value.Uint16Value() != 1000 {
		t.Errorf("Expected 1000 after assignment, got %d", value.Uint16Value())
	}

	value.Assign(uint16(0xffff))
	if value.Uint16Value() != 0xffff {
		t.Errorf("Expected 0xffff after assignment, got %d", value.Uint16Value())
	}
}

func TestUInt16Compare(t *testing.T) {
	tests := []struct {
		a, b     uint16
		expected int
	}{
		{0xffff, 0xfff0, 1},
		{0xf0f0, 0xf0f0, 0},
		{0x0fff, 0xfff0, -1},
	}

	for _, test := range tests {
		u1 := NewUInt16(test.a)
		u2 := NewUInt16(test.b)
		result := u1.CompareTo(u2)
		if result != test.expected {
			t.Errorf("UInt16(%d).CompareTo(UInt16(%d)) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestUInt16Serialize(t *testing.T) {
	// Test Put and Get
	value := NewUInt16(0x1234)
	buf := &bytes.Buffer{}
	
	err := value.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put UInt16: %v", err)
	}
	
	if buf.Len() != 2 {
		t.Errorf("Expected buffer length 2, got %d", buf.Len())
	}
	
	// Read back
	result := NewUInt16(0)
	err = result.Get(buf)
	if err != nil {
		t.Fatalf("Failed to get UInt16: %v", err)
	}
	
	if result.Uint16Value() != 0x1234 {
		t.Errorf("Expected 0x1234, got 0x%x", result.Uint16Value())
	}
}

func TestUInt16SerializeEndian(t *testing.T) {
	// Test little endian byte order
	value := NewUInt16(0x1234)
	buf := &bytes.Buffer{}
	
	err := value.Put(buf)
	if err != nil {
		t.Fatalf("Failed to put UInt16: %v", err)
	}
	
	data := buf.Bytes()
	if len(data) != 2 {
		t.Fatalf("Expected 2 bytes, got %d", len(data))
	}
	
	// Little endian: 0x1234 should be stored as [0x34, 0x12]
	if data[0] != 0x34 || data[1] != 0x12 {
		t.Errorf("Expected [0x34, 0x12], got [0x%02x, 0x%02x]", data[0], data[1])
	}
}