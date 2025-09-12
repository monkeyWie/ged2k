package protocol

import (
	"bytes"
	"fmt"
	"math"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
)

// BitField represents a bit field
type BitField struct {
	bytes []byte
	size  int
}

// NewBitField creates a new empty BitField
func NewBitField() *BitField {
	return &BitField{}
}

// NewBitFieldWithSize creates a new BitField with specified size
func NewBitFieldWithSize(bits int) *BitField {
	bf := &BitField{}
	bf.Resize(bits)
	return bf
}

// NewBitFieldWithValue creates a new BitField with specified size and value
func NewBitFieldWithValue(bits int, val bool) *BitField {
	bf := &BitField{}
	bf.ResizeWithValue(bits, val)
	return bf
}

// NewBitFieldFromBytes creates a new BitField from bytes
func NewBitFieldFromBytes(b []byte, bits int) *BitField {
	bf := &BitField{}
	bf.Assign(b, bits)
	return bf
}

// NewBitFieldCopy creates a copy of another BitField
func NewBitFieldCopy(rhs *BitField) *BitField {
	return NewBitFieldFromBytes(rhs.Bytes(), rhs.Size())
}

// bitsToBytes calculates number of bytes needed for given bits
func (bf *BitField) bitsToBytes(count int) int {
	return int(math.Ceil(float64(count) / 8.0))
}

// Assign assigns bytes and bits to the BitField
func (bf *BitField) Assign(b []byte, bits int) {
	bf.Resize(bits)
	bytesToCopy := bf.bitsToBytes(bits)
	if bytesToCopy > 0 {
		copy(bf.bytes, b[:bytesToCopy])
	}
	bf.clearTrailingBits()
}

// GetBit returns the bit at the specified index
func (bf *BitField) GetBit(index int) bool {
	if index < 0 || index >= bf.size {
		panic("index out of range")
	}
	return (bf.bytes[index/8] & (0x80 >> (index & 7))) != 0
}

// ClearBit clears the bit at the specified index
func (bf *BitField) ClearBit(index int) {
	if bf.bytes == nil || index < 0 || index >= bf.size {
		panic("index out of range")
	}
	bf.bytes[index/8] &= ^(0x80 >> (index & 7))
}

// SetBit sets the bit at the specified index
func (bf *BitField) SetBit(index int) {
	if bf.bytes == nil || index < 0 || index >= bf.size {
		panic("index out of range")
	}
	bf.bytes[index/8] |= (0x80 >> (index & 7))
}

// Size returns the size in bits
func (bf *BitField) Size() int {
	return bf.size
}

// Empty returns true if the BitField is empty
func (bf *BitField) Empty() bool {
	return bf.size == 0
}

// Bytes returns the underlying byte array
func (bf *BitField) Bytes() []byte {
	return bf.bytes
}

// Count returns the number of set bits
func (bf *BitField) Count() int {
	if bf.bytes == nil {
		return 0
	}

	// Bit count lookup table for nibbles (0-15)
	numBits := []int{
		0, 1, 1, 2, 1, 2, 2, 3,
		1, 2, 2, 3, 2, 3, 3, 4,
	}

	ret := 0
	numBytes := bf.size / 8
	
	// Count bits in complete bytes
	for i := 0; i < numBytes; i++ {
		ret += numBits[bf.bytes[i]&0xf] + numBits[bf.bytes[i]>>4]
	}

	// Count bits in the remaining partial byte
	rest := bf.size - numBytes*8
	for i := 0; i < rest; i++ {
		ret += int((bf.bytes[numBytes] >> (7 - i)) & 1)
	}

	return ret
}

// ResizeWithValue resizes the BitField with specified value for new bits
func (bf *BitField) ResizeWithValue(bits int, val bool) {
	s := bf.size
	b := bf.size & 7 // bits in last byte
	bf.Resize(bits)

	if s >= bf.size {
		return
	}

	oldSizeBytes := bf.bitsToBytes(s)
	newSizeBytes := bf.bitsToBytes(bf.size)

	if val {
		if oldSizeBytes != 0 && b != 0 {
			bf.bytes[oldSizeBytes-1] |= (0xff >> b)
		}
		if oldSizeBytes < newSizeBytes {
			for i := oldSizeBytes; i < newSizeBytes; i++ {
				bf.bytes[i] = 0xff
			}
		}
		bf.clearTrailingBits()
	} else {
		if oldSizeBytes < newSizeBytes {
			for i := oldSizeBytes; i < newSizeBytes; i++ {
				bf.bytes[i] = 0x00
			}
		}
	}
}

// SetAll sets all bits to 1
func (bf *BitField) SetAll() {
	for i := range bf.bytes {
		bf.bytes[i] = 0xff
	}
	bf.clearTrailingBits()
}

// ClearAll sets all bits to 0
func (bf *BitField) ClearAll() {
	for i := range bf.bytes {
		bf.bytes[i] = 0x00
	}
}

// Resize resizes the BitField
func (bf *BitField) Resize(bits int) {
	if bits < 0 {
		panic("bits cannot be negative")
	}
	
	b := bf.bitsToBytes(bits)
	newBytes := make([]byte, b)

	if bf.bytes != nil {
		copyLen := len(bf.bytes)
		if copyLen > len(newBytes) {
			copyLen = len(newBytes)
		}
		copy(newBytes, bf.bytes[:copyLen])
	}

	bf.bytes = newBytes
	bf.size = bits
	bf.clearTrailingBits()
}

// clearTrailingBits clears the tail bits in the last byte
func (bf *BitField) clearTrailingBits() {
	if (bf.size & 7) != 0 {
		bf.bytes[bf.bitsToBytes(bf.size)-1] &= 0xff << (8 - (bf.size & 7))
	}
}

// Equals checks if two BitFields are equal
func (bf *BitField) Equals(other *BitField) bool {
	if other == nil {
		return false
	}
	if bf.size != other.size {
		return false
	}
	for i := 0; i < bf.size; i++ {
		if bf.GetBit(i) != other.GetBit(i) {
			return false
		}
	}
	return true
}

// String returns string representation of the BitField
func (bf *BitField) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d[", bf.size))
	for i := 0; i < bf.size; i++ {
		if bf.GetBit(i) {
			buf.WriteString("1")
		} else {
			buf.WriteString("0")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

// Get implements Serializable interface - deserializes from buffer
func (bf *BitField) Get(src []byte, offset int) (int, error) {
	if len(src) < offset+2 {
		return 0, exception.NewJED2KExceptionWithMessage(exception.BUFFER_UNDERFLOW_EXCEPTION, "buffer too small for size")
	}
	
	size := int(uint16(src[offset])<<8 | uint16(src[offset+1]))
	offset += 2
	
	bytesNeeded := bf.bitsToBytes(size)
	if len(src) < offset+bytesNeeded {
		return 0, exception.NewJED2KExceptionWithMessage(exception.BUFFER_UNDERFLOW_EXCEPTION, "buffer too small for data")
	}
	
	temp := make([]byte, bytesNeeded)
	copy(temp, src[offset:offset+bytesNeeded])
	bf.Assign(temp, size)
	
	return offset + bytesNeeded, nil
}

// Put implements Serializable interface - serializes to buffer
func (bf *BitField) Put(dst []byte, offset int) (int, error) {
	if len(dst) < offset+2 {
		return 0, exception.NewJED2KExceptionWithMessage(exception.BUFFER_OVERFLOW_EXCEPTION, "buffer too small for size")
	}
	
	dst[offset] = byte(bf.size >> 8)
	dst[offset+1] = byte(bf.size)
	offset += 2
	
	if bf.bytes != nil {
		if len(dst) < offset+len(bf.bytes) {
			return 0, exception.NewJED2KExceptionWithMessage(exception.BUFFER_OVERFLOW_EXCEPTION, "buffer too small for data")
		}
		copy(dst[offset:], bf.bytes)
		offset += len(bf.bytes)
	}
	
	return offset, nil
}

// BytesCount returns the number of bytes needed for serialization
func (bf *BitField) BytesCount() int {
	bytesForData := 0
	if bf.bytes != nil {
		bytesForData = len(bf.bytes)
	}
	return 2 + bytesForData // 2 bytes for size + data bytes
}