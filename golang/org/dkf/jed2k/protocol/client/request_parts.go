package client

import (
	"bytes"
	"fmt"
	
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// BlockRequest represents a request for a file block range
type BlockRequest struct {
	StartOffset uint64
	EndOffset   uint64
}

// RequestParts64 packet to request file parts
type RequestParts64 struct {
	FileHash *protocol.Hash
	Requests []BlockRequest
}

func NewRequestParts64(hash *protocol.Hash) *RequestParts64 {
	return &RequestParts64{
		FileHash: hash,
		Requests: make([]BlockRequest, 0),
	}
}

// AddRequest adds a block request
func (r *RequestParts64) AddRequest(startOffset, endOffset uint64) {
	r.Requests = append(r.Requests, BlockRequest{
		StartOffset: startOffset,
		EndOffset:   endOffset,
	})
}

func (r *RequestParts64) Get(src *bytes.Buffer) error {
	return nil
}

func (r *RequestParts64) Put(dst *bytes.Buffer) error {
	// Write file hash
	if err := r.FileHash.Put(dst); err != nil {
		return fmt.Errorf("failed to write file hash: %v", err)
	}
	
	// Write each request
	for _, req := range r.Requests {
		startOffset := protocol.NewUInt64(req.StartOffset)
		if err := startOffset.Put(dst); err != nil {
			return fmt.Errorf("failed to write start offset: %v", err)
		}
		
		endOffset := protocol.NewUInt64(req.EndOffset)
		if err := endOffset.Put(dst); err != nil {
			return fmt.Errorf("failed to write end offset: %v", err)
		}
	}
	
	return nil
}

func (r *RequestParts64) BytesCount() int {
	return 16 + len(r.Requests)*16 // hash + (start_offset + end_offset) * count
}

func (r *RequestParts64) IsEmpty() bool {
	return len(r.Requests) == 0
}

// SendingPart64 packet containing requested file data
type SendingPart64 struct {
	FileHash     *protocol.Hash
	StartOffset  uint64
	EndOffset    uint64
	Data         []byte
}

func (s *SendingPart64) Get(src *bytes.Buffer) error {
	return nil
}

func (s *SendingPart64) Put(dst *bytes.Buffer) error {
	// Write file hash
	if err := s.FileHash.Put(dst); err != nil {
		return fmt.Errorf("failed to write file hash: %v", err)
	}
	
	// Write start offset
	startOffset := protocol.NewUInt64(s.StartOffset)
	if err := startOffset.Put(dst); err != nil {
		return fmt.Errorf("failed to write start offset: %v", err)
	}
	
	// Write end offset
	endOffset := protocol.NewUInt64(s.EndOffset)
	if err := endOffset.Put(dst); err != nil {
		return fmt.Errorf("failed to write end offset: %v", err)
	}
	
	// Write data
	if _, err := dst.Write(s.Data); err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	
	return nil
}

func (s *SendingPart64) BytesCount() int {
	return 16 + 8 + 8 + len(s.Data) // hash + start_offset + end_offset + data
}