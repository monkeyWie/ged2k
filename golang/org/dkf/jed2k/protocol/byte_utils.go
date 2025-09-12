package protocol

import (
	"io"
)

// ByteWriter is a simple writer that collects bytes into a slice
type ByteWriter struct {
	Data []byte
}

// Write implements io.Writer interface
func (bw *ByteWriter) Write(p []byte) (n int, err error) {
	bw.Data = append(bw.Data, p...)
	return len(p), nil
}

// Reset clears the internal buffer
func (bw *ByteWriter) Reset() {
	bw.Data = bw.Data[:0]
}

// Len returns the length of accumulated data
func (bw *ByteWriter) Len() int {
	return len(bw.Data)
}

// Bytes returns a copy of the accumulated data
func (bw *ByteWriter) Bytes() []byte {
	return append([]byte(nil), bw.Data...)
}

// ByteReader is a simple reader that reads from a byte slice
type ByteReader struct {
	Data   []byte
	offset int
}

// NewByteReader creates a new ByteReader
func NewByteReader(data []byte) *ByteReader {
	return &ByteReader{Data: data, offset: 0}
}

// Read implements io.Reader interface
func (br *ByteReader) Read(p []byte) (n int, err error) {
	if br.offset >= len(br.Data) {
		return 0, io.EOF
	}
	
	n = copy(p, br.Data[br.offset:])
	br.offset += n
	return n, nil
}

// ReadByte reads a single byte
func (br *ByteReader) ReadByte() (byte, error) {
	if br.offset >= len(br.Data) {
		return 0, io.EOF
	}
	
	b := br.Data[br.offset]
	br.offset++
	return b, nil
}

// Remaining returns the number of bytes remaining
func (br *ByteReader) Remaining() int {
	return len(br.Data) - br.offset
}

// Reset resets the reader to the beginning
func (br *ByteReader) Reset() {
	br.offset = 0
}

// Seek sets the read position
func (br *ByteReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		br.offset = int(offset)
	case io.SeekCurrent:
		br.offset += int(offset)
	case io.SeekEnd:
		br.offset = len(br.Data) + int(offset)
	}
	
	if br.offset < 0 {
		br.offset = 0
	} else if br.offset > len(br.Data) {
		br.offset = len(br.Data)
	}
	
	return int64(br.offset), nil
}