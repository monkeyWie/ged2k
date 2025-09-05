package client

import (
	"bytes"
	"fmt"
	
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// OP_HELLO packet for peer connection establishment
type Hello struct {
	UserHash    *protocol.Hash
	ClientID    *protocol.UInt32
	ListenPort  *protocol.UInt16
	Tags        []protocol.Tag
}

// NewHello creates a new Hello packet
func NewHello(userHash *protocol.Hash, clientID uint32, port uint16) *Hello {
	return &Hello{
		UserHash:   userHash,
		ClientID:   protocol.NewUInt32(clientID),
		ListenPort: protocol.NewUInt16(port),
		Tags:       make([]protocol.Tag, 0),
	}
}

func (h *Hello) Get(src *bytes.Buffer) error {
	return nil
}

func (h *Hello) Put(dst *bytes.Buffer) error {
	// Write user hash
	if err := h.UserHash.Put(dst); err != nil {
		return fmt.Errorf("failed to write user hash: %v", err)
	}
	
	// Write client ID
	if err := h.ClientID.Put(dst); err != nil {
		return fmt.Errorf("failed to write client ID: %v", err)
	}
	
	// Write listen port
	if err := h.ListenPort.Put(dst); err != nil {
		return fmt.Errorf("failed to write listen port: %v", err)
	}
	
	// Write number of tags
	tagCount := protocol.NewUInt32FromInt(len(h.Tags))
	if err := tagCount.Put(dst); err != nil {
		return fmt.Errorf("failed to write tag count: %v", err)
	}
	
	// Write tags
	for _, tag := range h.Tags {
		tagBuf := &bytes.Buffer{}
		if err := tag.Put(tagBuf); err != nil {
			return fmt.Errorf("failed to write tag: %v", err)
		}
		dst.Write(tagBuf.Bytes())
	}
	
	return nil
}

func (h *Hello) BytesCount() int {
	count := 16 + 4 + 2 + 4 // hash + client_id + port + tag_count
	for _, tag := range h.Tags {
		count += tag.BytesCount()
	}
	return count
}

// HelloAnswer packet response
type HelloAnswer struct {
	UserHash    *protocol.Hash
	ClientID    *protocol.UInt32
	ListenPort  *protocol.UInt16
	Tags        []protocol.Tag
	ServerIP    *protocol.UInt32
	ServerPort  *protocol.UInt16
}

func (h *HelloAnswer) Get(src *bytes.Buffer) error {
	return nil
}

func (h *HelloAnswer) Put(dst *bytes.Buffer) error {
	// Similar to Hello, but with server info
	if err := h.UserHash.Put(dst); err != nil {
		return fmt.Errorf("failed to write user hash: %v", err)
	}
	
	if err := h.ClientID.Put(dst); err != nil {
		return fmt.Errorf("failed to write client ID: %v", err)
	}
	
	if err := h.ListenPort.Put(dst); err != nil {
		return fmt.Errorf("failed to write listen port: %v", err)
	}
	
	tagCount := protocol.NewUInt32FromInt(len(h.Tags))
	if err := tagCount.Put(dst); err != nil {
		return fmt.Errorf("failed to write tag count: %v", err)
	}
	
	for _, tag := range h.Tags {
		tagBuf := &bytes.Buffer{}
		if err := tag.Put(tagBuf); err != nil {
			return fmt.Errorf("failed to write tag: %v", err)
		}
		dst.Write(tagBuf.Bytes())
	}
	
	if err := h.ServerIP.Put(dst); err != nil {
		return fmt.Errorf("failed to write server IP: %v", err)
	}
	
	if err := h.ServerPort.Put(dst); err != nil {
		return fmt.Errorf("failed to write server port: %v", err)
	}
	
	return nil
}

func (h *HelloAnswer) BytesCount() int {
	count := 16 + 4 + 2 + 4 + 4 + 2 // hash + client_id + port + tag_count + server_ip + server_port
	for _, tag := range h.Tags {
		count += tag.BytesCount()
	}
	return count
}