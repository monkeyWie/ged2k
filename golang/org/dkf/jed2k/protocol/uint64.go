package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// UInt64 represents a 64-bit unsigned integer
type UInt64 struct {
	value uint64
}

const (
	UInt64MinValue = 0x0000000000000000
	UInt64MaxValue = 0xffffffffffffffff
)

// NewUInt64 creates a new UInt64 with the given value
func NewUInt64(value uint64) *UInt64 {
	return &UInt64{value: value}
}

// NewUInt64FromInt creates a new UInt64 from an int value
func NewUInt64FromInt(value int) *UInt64 {
	return &UInt64{value: uint64(value)}
}

// Get deserializes from buffer (little endian)
func (u *UInt64) Get(src *bytes.Buffer) error {
	if src.Len() < 8 {
		return fmt.Errorf("buffer underflow: need 8 bytes, have %d", src.Len())
	}
	return binary.Read(src, binary.LittleEndian, &u.value)
}

// Put serializes to buffer (little endian)
func (u *UInt64) Put(dst *bytes.Buffer) error {
	return binary.Write(dst, binary.LittleEndian, u.value)
}

// BytesCount returns the number of bytes this type uses
func (u *UInt64) BytesCount() int {
	return 8
}

// Value accessors
func (u *UInt64) Uint8Value() uint8   { return uint8(u.value) }
func (u *UInt64) Uint16Value() uint16 { return uint16(u.value) }
func (u *UInt64) Uint32Value() uint32 { return uint32(u.value) }
func (u *UInt64) Uint64Value() uint64 { return u.value }
func (u *UInt64) Int8Value() int8     { return int8(u.value) }
func (u *UInt64) Int16Value() int16   { return int16(u.value) }
func (u *UInt64) Int32Value() int32   { return int32(u.value) }
func (u *UInt64) Int64Value() int64   { return int64(u.value) }
func (u *UInt64) Float32Value() float32 { return float32(u.value) }
func (u *UInt64) Float64Value() float64 { return float64(u.value) }

// Assign sets the value from various types
func (u *UInt64) Assign(value interface{}) UNumber {
	switch v := value.(type) {
	case uint8:
		u.value = uint64(v)
	case int:
		u.value = uint64(v)
	case int8:
		u.value = uint64(v)
	case uint16:
		u.value = uint64(v)
	case int16:
		u.value = uint64(v)
	case uint32:
		u.value = uint64(v)
	case int32:
		u.value = uint64(v)
	case uint64:
		u.value = v
	case int64:
		u.value = uint64(v)
	}
	return u
}

// Equals checks equality
func (u *UInt64) Equals(other *UInt64) bool {
	return u.value == other.value
}

// CompareTo compares with another UInt64
func (u *UInt64) CompareTo(other *UInt64) int {
	if u.value < other.value {
		return -1
	} else if u.value > other.value {
		return 1
	}
	return 0
}

// String returns string representation
func (u *UInt64) String() string {
	return fmt.Sprintf("uint64{%d}", u.value)
}