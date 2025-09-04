package protocol

import (
	"bytes"
)

// Serializable interface for objects that can be serialized to/from binary data
type Serializable interface {
	Get(src *bytes.Buffer) error
	Put(dst *bytes.Buffer) error
	BytesCount() int
}

// UNumber is the base interface for unsigned number types
type UNumber interface {
	Serializable
	Uint8Value() uint8
	Uint16Value() uint16  
	Uint32Value() uint32
	Uint64Value() uint64
	Int8Value() int8
	Int16Value() int16
	Int32Value() int32
	Int64Value() int64
	Float32Value() float32
	Float64Value() float64
	Assign(value interface{}) UNumber
}