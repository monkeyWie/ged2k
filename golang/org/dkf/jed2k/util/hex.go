package util

import (
	"fmt"
	"strconv"
	"strings"
)

// Hex provides hexadecimal conversion utilities
type Hex struct{}

var (
	// DIGITS_LOWER used for lowercase hex output
	DIGITS_LOWER = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
	
	// DIGITS_UPPER used for uppercase hex output  
	DIGITS_UPPER = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}
)

// Decode converts hex characters to bytes
func Decode(data []rune) ([]byte, error) {
	length := len(data)
	if length%2 != 0 {
		return nil, fmt.Errorf("hex data must have even length")
	}
	
	out := make([]byte, length>>1)
	
	// Two characters form the hex value
	for i, j := 0, 0; j < length; i++ {
		f, err := toDigit(data[j], j)
		if err != nil {
			return nil, err
		}
		f <<= 4
		j++
		
		digit, err := toDigit(data[j], j)
		if err != nil {
			return nil, err
		}
		f |= digit
		j++
		
		out[i] = byte(f & 0xFF)
	}
	
	return out, nil
}

// DecodeString converts hex string to bytes
func DecodeString(data string) ([]byte, error) {
	return Decode([]rune(data))
}

// Encode converts bytes to hex string (lowercase)
func Encode(data []byte) string {
	return string(EncodeToChars(data, true))
}

// EncodeToChars converts bytes to hex characters
func EncodeToChars(data []byte, toLowerCase bool) []rune {
	digits := DIGITS_UPPER
	if toLowerCase {
		digits = DIGITS_LOWER
	}
	return encodeWithDigits(data, digits)
}

// encodeWithDigits converts bytes using specific digit set
func encodeWithDigits(data []byte, toDigits []rune) []rune {
	length := len(data)
	out := make([]rune, length<<1)
	
	// Two characters form the hex value
	for i, j := 0, 0; i < length; i++ {
		out[j] = toDigits[(0xF0&data[i])>>4]
		j++
		out[j] = toDigits[0x0F&data[i]]
		j++
	}
	
	return out
}

// toDigit converts hex character to digit
func toDigit(ch rune, index int) (int, error) {
	// Convert rune to string and parse as hex
	chStr := string(ch)
	digit, err := strconv.ParseInt(chStr, 16, 32)
	if err != nil {
		return -1, fmt.Errorf("illegal hexadecimal character %c at index %d", ch, index)
	}
	return int(digit), nil
}

// HexDump provides hex dump utilities
type HexDump struct{}

// Dump creates a hex dump of the given bytes
func (h *HexDump) Dump(data []byte) string {
	var result strings.Builder
	
	for i := 0; i < len(data); i += 16 {
		// Address
		result.WriteString(fmt.Sprintf("%08x: ", i))
		
		// Hex bytes
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				result.WriteString(fmt.Sprintf("%02x ", data[i+j]))
			} else {
				result.WriteString("   ")
			}
			
			if j == 7 {
				result.WriteString(" ")
			}
		}
		
		// ASCII representation
		result.WriteString(" |")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				result.WriteByte(b)
			} else {
				result.WriteString(".")
			}
		}
		result.WriteString("|\n")
	}
	
	return result.String()
}