package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
)

// Server represents an ED2K server
type Server struct {
	IP          uint32 `json:"IP"`
	Port        uint16 `json:"Port"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Ping        uint32 `json:"Ping"`
	MaxUsers    uint32 `json:"MaxUsers"`
	Files       uint32 `json:"Files"`
	Users       uint32 `json:"Users"`
	LowIDUsers  uint32 `json:"LowIDUsers"`
	Version     string `json:"Version"`
	UDPFlags    uint32 `json:"UDPFlags"`
	Tags        []Tag  `json:"Tags"`
}

// Tag represents a server information tag
type Tag struct {
	Type  byte        `json:"Type"`
	Name  string      `json:"Name"`
	Value interface{} `json:"Value"`
}

// NewStringTag creates a new string tag
func NewStringTag(name string, value string) *Tag {
	return &Tag{
		Type:  TagTypeString,
		Name:  name,
		Value: value,
	}
}

// NewUInt32Tag creates a new uint32 tag
func NewUInt32Tag(name string, value uint32) *Tag {
	return &Tag{
		Type:  TagTypeUInt32,
		Name:  name,
		Value: value,
	}
}

func (t *Tag) Get(src *bytes.Buffer) error {
	return nil
}

func (t *Tag) Put(dst *bytes.Buffer) error {
	// Write tag type
	if err := binary.Write(dst, binary.LittleEndian, t.Type); err != nil {
		return fmt.Errorf("failed to write tag type: %v", err)
	}
	
	// Write name length and name
	nameBytes := []byte(t.Name)
	nameLen := uint16(len(nameBytes))
	if err := binary.Write(dst, binary.LittleEndian, nameLen); err != nil {
		return fmt.Errorf("failed to write name length: %v", err)
	}
	
	if _, err := dst.Write(nameBytes); err != nil {
		return fmt.Errorf("failed to write name: %v", err)
	}
	
	// Write value based on type
	switch t.Type {
	case TagTypeString:
		valueBytes := []byte(t.Value.(string))
		valueLen := uint16(len(valueBytes))
		if err := binary.Write(dst, binary.LittleEndian, valueLen); err != nil {
			return fmt.Errorf("failed to write string value length: %v", err)
		}
		if _, err := dst.Write(valueBytes); err != nil {
			return fmt.Errorf("failed to write string value: %v", err)
		}
	case TagTypeUInt32:
		if err := binary.Write(dst, binary.LittleEndian, t.Value.(uint32)); err != nil {
			return fmt.Errorf("failed to write uint32 value: %v", err)
		}
	case TagTypeUInt16:
		if err := binary.Write(dst, binary.LittleEndian, t.Value.(uint16)); err != nil {
			return fmt.Errorf("failed to write uint16 value: %v", err)
		}
	case TagTypeUInt8:
		if err := binary.Write(dst, binary.LittleEndian, t.Value.(uint8)); err != nil {
			return fmt.Errorf("failed to write uint8 value: %v", err)
		}
	default:
		return fmt.Errorf("unsupported tag type: %d", t.Type)
	}
	
	return nil
}

func (t *Tag) BytesCount() int {
	count := 1 + 2 + len(t.Name) // type + name_len + name
	
	switch t.Type {
	case TagTypeString:
		count += 2 + len(t.Value.(string)) // value_len + value
	case TagTypeUInt32:
		count += 4
	case TagTypeUInt16:
		count += 2
	case TagTypeUInt8:
		count += 1
	}
	
	return count
}

// Server tag types
const (
	TagServerName     byte = 0x01
	TagDescription    byte = 0x0B
	TagPing           byte = 0x0C
	TagFail           byte = 0x0D
	TagPreference     byte = 0x0E
	TagDynIP          byte = 0x85
	TagLastPing       byte = 0x86
	TagMaxUsers       byte = 0x87
	TagSoftFiles      byte = 0x88
	TagHardFiles      byte = 0x89
	TagLastDNSResolve byte = 0x90
	TagVersion        byte = 0x91
	TagUDPFlags       byte = 0x92
	TagAuxPortsList   byte = 0x93
	TagLowIDUsers     byte = 0x94
)

// Tag value types
const (
	TagTypeHash    byte = 0x01
	TagTypeString  byte = 0x02
	TagTypeUInt32  byte = 0x03
	TagTypeFloat32 byte = 0x04
	TagTypeBool    byte = 0x05
	TagTypeBoolArr byte = 0x06
	TagTypeBlob    byte = 0x07
	TagTypeUInt16  byte = 0x08
	TagTypeUInt8   byte = 0x09
	TagTypeUInt64  byte = 0x0B
)

// ServerMet represents a server.met file
type ServerMet struct {
	Version byte     `json:"Version"`
	Servers []Server `json:"Servers"`
}

// ParseServerMet parses a server.met file
func ParseServerMet(data []byte) (*ServerMet, error) {
	if len(data) < 5 {
		return nil, errors.New("invalid server.met file: too short")
	}
	
	serverMet := &ServerMet{}
	offset := 0
	
	// Parse version
	serverMet.Version = data[offset]
	offset++
	
	// Accept common server.met versions
	if serverMet.Version != 0x0E && serverMet.Version != 0x0F && serverMet.Version != 0xE0 {
		return nil, fmt.Errorf("invalid server.met version: 0x%02X", serverMet.Version)
	}
	
	// Parse number of servers
	if len(data) < offset+4 {
		return nil, errors.New("invalid server.met file: missing server count")
	}
	numServers := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	
	serverMet.Servers = make([]Server, 0, numServers)
	
	// Parse each server
	for i := uint32(0); i < numServers; i++ {
		server, newOffset, err := parseServerEntry(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing server %d: %v", i, err)
		}
		serverMet.Servers = append(serverMet.Servers, server)
		offset = newOffset
	}
	
	return serverMet, nil
}

// parseServerEntry parses a single server entry from the data
func parseServerEntry(data []byte, offset int) (Server, int, error) {
	server := Server{}
	
	// Parse IP and port (6 bytes)
	if len(data) < offset+6 {
		return server, offset, errors.New("insufficient data for server IP/port")
	}
	
	server.IP = binary.LittleEndian.Uint32(data[offset : offset+4])
	server.Port = binary.LittleEndian.Uint16(data[offset+4 : offset+6])
	offset += 6
	
	// Parse number of tags
	if len(data) < offset+4 {
		return server, offset, errors.New("insufficient data for tag count")
	}
	numTags := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	
	// Parse tags
	for i := uint32(0); i < numTags; i++ {
		tag, newOffset, err := parseTag(data, offset)
		if err != nil {
			return server, offset, fmt.Errorf("error parsing tag %d: %v", i, err)
		}
		
		// Extract common server properties from tags
		switch tag.Type {
		case TagServerName:
			if str, ok := tag.Value.(string); ok {
				server.Name = str
			}
		case TagDescription:
			if str, ok := tag.Value.(string); ok {
				server.Description = str
			}
		case TagPing:
			if val, ok := tag.Value.(uint32); ok {
				server.Ping = val
			}
		case TagMaxUsers:
			if val, ok := tag.Value.(uint32); ok {
				server.MaxUsers = val
			}
		case TagSoftFiles:
			if val, ok := tag.Value.(uint32); ok {
				server.Files = val
			}
		case TagLowIDUsers:
			if val, ok := tag.Value.(uint32); ok {
				server.LowIDUsers = val
			}
		case TagVersion:
			if str, ok := tag.Value.(string); ok {
				server.Version = str
			}
		case TagUDPFlags:
			if val, ok := tag.Value.(uint32); ok {
				server.UDPFlags = val
			}
		}
		
		server.Tags = append(server.Tags, tag)
		offset = newOffset
	}
	
	return server, offset, nil
}

// parseTag parses a single tag from the data
func parseTag(data []byte, offset int) (Tag, int, error) {
	tag := Tag{}
	
	if len(data) < offset+1 {
		return tag, offset, errors.New("insufficient data for tag type")
	}
	
	// Parse tag value type
	tagValueType := data[offset]
	offset++
	
	if len(data) < offset+2 {
		return tag, offset, errors.New("insufficient data for tag name length")
	}
	nameLen := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2
	
	if nameLen > 1 {
		if len(data) < offset+int(nameLen) {
			return tag, offset, errors.New("insufficient data for tag name")
		}
		tag.Name = string(data[offset : offset+int(nameLen)])
		tag.Type = 0  // Tags with names have Type=0
		offset += int(nameLen)
	} else if nameLen == 1 {
		// Parse tag ID (1 byte) - this is actually the common case for server tags
		if len(data) < offset+1 {
			return tag, offset, errors.New("insufficient data for tag ID")
		}
		tag.Type = data[offset]
		tag.Name = ""  // Tags with IDs have empty names
		offset++
	} else {
		// nameLen == 0, which should not happen in normal server.met files
		tag.Type = 0
		tag.Name = ""
	}
	
	// Parse tag value based on type
	switch tagValueType {
	case TagTypeString:
		if len(data) < offset+2 {
			return tag, offset, errors.New("insufficient data for string length")
		}
		strLen := binary.LittleEndian.Uint16(data[offset : offset+2])
		offset += 2
		if len(data) < offset+int(strLen) {
			return tag, offset, errors.New("insufficient data for string value")
		}
		tag.Value = string(data[offset : offset+int(strLen)])
		offset += int(strLen)
		
	case TagTypeUInt32:
		if len(data) < offset+4 {
			return tag, offset, errors.New("insufficient data for uint32 value")
		}
		tag.Value = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
		
	case TagTypeUInt16:
		if len(data) < offset+2 {
			return tag, offset, errors.New("insufficient data for uint16 value")
		}
		tag.Value = binary.LittleEndian.Uint16(data[offset : offset+2])
		offset += 2
		
	case TagTypeUInt8:
		if len(data) < offset+1 {
			return tag, offset, errors.New("insufficient data for uint8 value")
		}
		tag.Value = data[offset]
		offset++
		
	case TagTypeFloat32:
		if len(data) < offset+4 {
			return tag, offset, errors.New("insufficient data for float32 value")
		}
		bits := binary.LittleEndian.Uint32(data[offset : offset+4])
		tag.Value = math.Float32frombits(bits)
		offset += 4
		
	default:
		return tag, offset, fmt.Errorf("unsupported tag type: 0x%02X", tagValueType)
	}
	
	return tag, offset, nil
}

// DownloadServerMet downloads and parses a server.met file from URL
func DownloadServerMet(url string) (*ServerMet, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download server.met: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download server.met: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read server.met data: %v", err)
	}

	return ParseServerMet(data)
}