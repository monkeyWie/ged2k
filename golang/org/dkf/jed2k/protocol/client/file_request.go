package client

import (
	"bytes"
	"fmt"
	
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// FileRequest packet to request a file
type FileRequest struct {
	FileHash *protocol.Hash
}

func NewFileRequest(hash *protocol.Hash) *FileRequest {
	return &FileRequest{
		FileHash: hash,
	}
}

func (f *FileRequest) Get(src *bytes.Buffer) error {
	return nil
}

func (f *FileRequest) Put(dst *bytes.Buffer) error {
	return f.FileHash.Put(dst)
}

func (f *FileRequest) BytesCount() int {
	return 16 // MD4 hash is 16 bytes
}

// FileAnswer packet response
type FileAnswer struct {
	FileHash     *protocol.Hash
	FileName     string
	Tags         []protocol.Tag
}

func (f *FileAnswer) Get(src *bytes.Buffer) error {
	return nil
}

func (f *FileAnswer) Put(dst *bytes.Buffer) error {
	if err := f.FileHash.Put(dst); err != nil {
		return fmt.Errorf("failed to write file hash: %v", err)
	}
	
	// Write filename length and filename
	nameBytes := []byte(f.FileName)
	nameLen := protocol.NewUInt16FromInt(len(nameBytes))
	if err := nameLen.Put(dst); err != nil {
		return fmt.Errorf("failed to write filename length: %v", err)
	}
	
	if _, err := dst.Write(nameBytes); err != nil {
		return fmt.Errorf("failed to write filename: %v", err)
	}
	
	// Write number of tags
	tagCount := protocol.NewUInt32FromInt(len(f.Tags))
	if err := tagCount.Put(dst); err != nil {
		return fmt.Errorf("failed to write tag count: %v", err)
	}
	
	// Write tags
	for _, tag := range f.Tags {
		tagBuf := &bytes.Buffer{}
		if err := tag.Put(tagBuf); err != nil {
			return fmt.Errorf("failed to write tag: %v", err)
		}
		dst.Write(tagBuf.Bytes())
	}
	
	return nil
}

func (f *FileAnswer) BytesCount() int {
	count := 16 + 2 + len(f.FileName) + 4 // hash + filename_len + filename + tag_count
	for _, tag := range f.Tags {
		count += tag.BytesCount()
	}
	return count
}