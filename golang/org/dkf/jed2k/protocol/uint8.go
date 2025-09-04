package protocol

import (
	"bytes"
	"fmt"
)

// UInt8 represents an 8-bit unsigned integer
type UInt8 struct {
	value uint8
}

const (
	UInt8MinValue = 0x00
	UInt8MaxValue = 0xff
)

// NewUInt8 creates a new UInt8 with the given value
func NewUInt8(value uint8) *UInt8 {
	return &UInt8{value: value}
}

// NewUInt8FromInt creates a new UInt8 from an int value
func NewUInt8FromInt(value int) *UInt8 {
	return &UInt8{value: uint8(value)}
}

// Get deserializes from buffer
func (u *UInt8) Get(src *bytes.Buffer) error {
	if src.Len() < 1 {
		return fmt.Errorf("buffer underflow: need 1 byte, have %d", src.Len())
	}
	var err error
	u.value, err = src.ReadByte()
	return err
}

// Put serializes to buffer
func (u *UInt8) Put(dst *bytes.Buffer) error {
	return dst.WriteByte(u.value)
}

// BytesCount returns the number of bytes this type uses
func (u *UInt8) BytesCount() int {
	return 1
}

// Value accessors
func (u *UInt8) Uint8Value() uint8   { return u.value }
func (u *UInt8) Uint16Value() uint16 { return uint16(u.value) }
func (u *UInt8) Uint32Value() uint32 { return uint32(u.value) }
func (u *UInt8) Uint64Value() uint64 { return uint64(u.value) }
func (u *UInt8) Int8Value() int8     { return int8(u.value) }
func (u *UInt8) Int16Value() int16   { return int16(u.value) }
func (u *UInt8) Int32Value() int32   { return int32(u.value) }
func (u *UInt8) Int64Value() int64   { return int64(u.value) }
func (u *UInt8) Float32Value() float32 { return float32(u.value) }
func (u *UInt8) Float64Value() float64 { return float64(u.value) }

// Assign sets the value from various types
func (u *UInt8) Assign(value interface{}) UNumber {
	switch v := value.(type) {
	case uint8:
		u.value = v
	case int:
		u.value = uint8(v)
	case int8:
		u.value = uint8(v)
	case uint16:
		u.value = uint8(v)
	case int16:
		u.value = uint8(v)
	case uint32:
		u.value = uint8(v)
	case int32:
		u.value = uint8(v)
	case uint64:
		u.value = uint8(v)
	case int64:
		u.value = uint8(v)
	}
	return u
}

// Equals checks equality
func (u *UInt8) Equals(other *UInt8) bool {
	return u.value == other.value
}

// CompareTo compares with another UInt8
func (u *UInt8) CompareTo(other *UInt8) int {
	if u.value < other.value {
		return -1
	} else if u.value > other.value {
		return 1
	}
	return 0
}

// String returns string representation
func (u *UInt8) String() string {
	return fmt.Sprintf("uint8{%d}", u.value)
}