package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// UInt16 represents a 16-bit unsigned integer
type UInt16 struct {
	value uint16
}

const (
	UInt16MinValue = 0x0000
	UInt16MaxValue = 0xffff
)

// NewUInt16 creates a new UInt16 with the given value
func NewUInt16(value uint16) *UInt16 {
	return &UInt16{value: value}
}

// NewUInt16FromInt creates a new UInt16 from an int value
func NewUInt16FromInt(value int) *UInt16 {
	return &UInt16{value: uint16(value)}
}

// Get deserializes from buffer (little endian)
func (u *UInt16) Get(src *bytes.Buffer) error {
	if src.Len() < 2 {
		return fmt.Errorf("buffer underflow: need 2 bytes, have %d", src.Len())
	}
	return binary.Read(src, binary.LittleEndian, &u.value)
}

// Put serializes to buffer (little endian)
func (u *UInt16) Put(dst *bytes.Buffer) error {
	return binary.Write(dst, binary.LittleEndian, u.value)
}

// BytesCount returns the number of bytes this type uses
func (u *UInt16) BytesCount() int {
	return 2
}

// Value accessors
func (u *UInt16) Uint8Value() uint8   { return uint8(u.value) }
func (u *UInt16) Uint16Value() uint16 { return u.value }
func (u *UInt16) Uint32Value() uint32 { return uint32(u.value) }
func (u *UInt16) Uint64Value() uint64 { return uint64(u.value) }
func (u *UInt16) Int8Value() int8     { return int8(u.value) }
func (u *UInt16) Int16Value() int16   { return int16(u.value) }
func (u *UInt16) Int32Value() int32   { return int32(u.value) }
func (u *UInt16) Int64Value() int64   { return int64(u.value) }
func (u *UInt16) Float32Value() float32 { return float32(u.value) }
func (u *UInt16) Float64Value() float64 { return float64(u.value) }

// Assign sets the value from various types
func (u *UInt16) Assign(value interface{}) UNumber {
	switch v := value.(type) {
	case uint8:
		u.value = uint16(v)
	case int:
		u.value = uint16(v)
	case int8:
		u.value = uint16(v)
	case uint16:
		u.value = v
	case int16:
		u.value = uint16(v)
	case uint32:
		u.value = uint16(v)
	case int32:
		u.value = uint16(v)
	case uint64:
		u.value = uint16(v)
	case int64:
		u.value = uint16(v)
	}
	return u
}

// Equals checks equality
func (u *UInt16) Equals(other *UInt16) bool {
	return u.value == other.value
}

// CompareTo compares with another UInt16
func (u *UInt16) CompareTo(other *UInt16) int {
	if u.value < other.value {
		return -1
	} else if u.value > other.value {
		return 1
	}
	return 0
}

// String returns string representation
func (u *UInt16) String() string {
	return fmt.Sprintf("uint16{%d}", u.value)
}