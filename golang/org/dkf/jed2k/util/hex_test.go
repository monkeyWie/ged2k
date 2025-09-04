package util

import (
	"testing"
)

func TestHexEncode(t *testing.T) {
	// Test basic encoding
	data := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	expected := "0123456789abcdef"
	
	result := Encode(data)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
	
	// Test uppercase encoding
	upperChars := EncodeToChars(data, false)
	upperResult := string(upperChars)
	expectedUpper := "0123456789ABCDEF"
	if upperResult != expectedUpper {
		t.Errorf("Expected %s, got %s", expectedUpper, upperResult)
	}
	
	// Test empty data
	emptyResult := Encode([]byte{})
	if emptyResult != "" {
		t.Errorf("Empty data should encode to empty string, got %s", emptyResult)
	}
}

func TestHexDecode(t *testing.T) {
	// Test basic decoding
	hexStr := "0123456789abcdef"
	expected := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	
	result, err := DecodeString(hexStr)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Byte %d: expected %02x, got %02x", i, expected[i], b)
		}
	}
	
	// Test uppercase decoding
	upperHexStr := "0123456789ABCDEF"
	result2, err := DecodeString(upperHexStr)
	if err != nil {
		t.Fatalf("Failed to decode uppercase hex string: %v", err)
	}
	
	for i, b := range result2 {
		if b != expected[i] {
			t.Errorf("Uppercase byte %d: expected %02x, got %02x", i, expected[i], b)
		}
	}
	
	// Test empty string
	emptyResult, err := DecodeString("")
	if err != nil {
		t.Fatalf("Failed to decode empty string: %v", err)
	}
	
	if len(emptyResult) != 0 {
		t.Error("Empty string should decode to empty byte array")
	}
	
	// Test odd length (should fail)
	_, err = DecodeString("123")
	if err == nil {
		t.Error("Odd length hex string should fail")
	}
	
	// Test invalid characters
	_, err = DecodeString("GG")
	if err == nil {
		t.Error("Invalid hex characters should fail")
	}
}

func TestHexRoundTrip(t *testing.T) {
	// Test round trip: data -> encode -> decode -> should equal original
	original := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	
	encoded := Encode(original)
	decoded, err := DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode encoded data: %v", err)
	}
	
	if len(decoded) != len(original) {
		t.Errorf("Length mismatch: original %d, decoded %d", len(original), len(decoded))
	}
	
	for i, b := range decoded {
		if b != original[i] {
			t.Errorf("Round trip failed at byte %d: original %02x, decoded %02x", i, original[i], b)
		}
	}
}

func TestHexDump(t *testing.T) {
	hexDump := &HexDump{}
	
	// Test with simple data
	data := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x21} // "Hello World!"
	
	result := hexDump.Dump(data)
	if len(result) == 0 {
		t.Error("Hex dump should not be empty")
	}
	
	// Should contain hex representation
	if !containsString(result, "48 65 6c 6c 6f") {
		t.Error("Hex dump should contain hex bytes")
	}
	
	// Should contain ASCII representation
	if !containsString(result, "Hello") {
		t.Error("Hex dump should contain ASCII representation")
	}
	
	// Test with empty data
	emptyResult := hexDump.Dump([]byte{})
	if len(emptyResult) != 0 {
		t.Error("Empty data should produce empty hex dump")
	}
	
	// Test with longer data (multiple lines)
	longData := make([]byte, 32)
	for i := range longData {
		longData[i] = byte(i)
	}
	
	longResult := hexDump.Dump(longData)
	if len(longResult) == 0 {
		t.Error("Long data hex dump should not be empty")
	}
	
	// Should contain multiple lines
	lines := countLines(longResult)
	if lines < 2 {
		t.Error("Long data should produce multiple lines")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// Simple substring search
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Count lines in string
func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}