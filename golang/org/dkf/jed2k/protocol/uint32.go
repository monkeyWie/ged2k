package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// UInt32 represents a 32-bit unsigned integer
type UInt32 struct {
	value uint32
}

const (
	UInt32MinValue = 0x00000000
	UInt32MaxValue = 0xffffffff
)

// NewUInt32 creates a new UInt32 with the given value
func NewUInt32(value uint32) *UInt32 {
	return &UInt32{value: value}
}

// NewUInt32FromInt creates a new UInt32 from an int value
func NewUInt32FromInt(value int) *UInt32 {
	return &UInt32{value: uint32(value)}
}

// Get deserializes from buffer (little endian)
func (u *UInt32) Get(src *bytes.Buffer) error {
	if src.Len() < 4 {
		return fmt.Errorf("buffer underflow: need 4 bytes, have %d", src.Len())
	}
	return binary.Read(src, binary.LittleEndian, &u.value)
}

// Put serializes to buffer (little endian)
func (u *UInt32) Put(dst *bytes.Buffer) error {
	return binary.Write(dst, binary.LittleEndian, u.value)
}

// BytesCount returns the number of bytes this type uses
func (u *UInt32) BytesCount() int {
	return 4
}

// Value accessors
func (u *UInt32) Uint8Value() uint8   { return uint8(u.value) }
func (u *UInt32) Uint16Value() uint16 { return uint16(u.value) }
func (u *UInt32) Uint32Value() uint32 { return u.value }
func (u *UInt32) Uint64Value() uint64 { return uint64(u.value) }
func (u *UInt32) Int8Value() int8     { return int8(u.value) }
func (u *UInt32) Int16Value() int16   { return int16(u.value) }
func (u *UInt32) Int32Value() int32   { return int32(u.value) }
func (u *UInt32) Int64Value() int64   { return int64(u.value) }
func (u *UInt32) Float32Value() float32 { return float32(u.value) }
func (u *UInt32) Float64Value() float64 { return float64(u.value) }

// Assign sets the value from various types
func (u *UInt32) Assign(value interface{}) UNumber {
	switch v := value.(type) {
	case uint8:
		u.value = uint32(v)
	case int:
		u.value = uint32(v)
	case int8:
		u.value = uint32(v)
	case uint16:
		u.value = uint32(v)
	case int16:
		u.value = uint32(v)
	case uint32:
		u.value = v
	case int32:
		u.value = uint32(v)
	case uint64:
		u.value = uint32(v)
	case int64:
		u.value = uint32(v)
	}
	return u
}

// Equals checks equality
func (u *UInt32) Equals(other *UInt32) bool {
	return u.value == other.value
}

// CompareTo compares with another UInt32
func (u *UInt32) CompareTo(other *UInt32) int {
	if u.value < other.value {
		return -1
	} else if u.value > other.value {
		return 1
	}
	return 0
}

// String returns string representation
func (u *UInt32) String() string {
	return fmt.Sprintf("uint32{%d}", u.value)
}