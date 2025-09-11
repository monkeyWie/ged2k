# ED2K Protocol Specification (RFC Draft)

**Network Working Group:** eDonkey2000/eMule Protocol Implementation  
**Request for Comments:** ED2K-PROTO-001  
**Category:** Standards Track  
**Date:** December 2024  
**Based on:** Java ged2k implementation and eMule protocol analysis

## Abstract

This document describes the complete eDonkey2000 (ed2k) peer-to-peer file sharing protocol, including client-server communication, peer-to-peer transfers, Kademlia DHT integration, and file metadata formats. The protocol enables distributed file sharing across a decentralized network of servers and peers.

## Table of Contents

1. [Introduction](#introduction)
2. [ED2K URI Specification](#ed2k-uri-specification)
3. [Protocol Fundamentals](#protocol-fundamentals)
4. [Client-Server Communication](#client-server-communication)
5. [Client-Client Communication](#client-client-communication)
6. [Kademlia Network Protocol](#kademlia-network-protocol)
7. [High ID and Low ID System](#high-id-and-low-id-system)
8. [Server.met File Specification](#servermet-file-specification)
9. [Nodes.dat File Specification](#nodesdat-file-specification)
10. [Security Considerations](#security-considerations)
11. [Implementation Guidelines](#implementation-guidelines)

---

## 1. Introduction

The eDonkey2000 protocol is a peer-to-peer file sharing protocol that combines centralized server coordination with direct peer-to-peer transfers. The protocol uses MD4 hashing for file identification and integrity verification.

### 1.1 Protocol Architecture

The ed2k network operates on a hybrid architecture:
- **Servers**: Provide centralized indexing and peer discovery
- **Clients**: Share files directly with other clients
- **Kademlia DHT**: Provides decentralized peer discovery and file indexing

### 1.2 Key Features

- File identification using MD4 hashes
- Chunk-based downloading with integrity verification
- Multi-source downloading from multiple peers
- Server-assisted and DHT-based peer discovery
- Resume capability for interrupted downloads

---

## 2. ED2K URI Specification

ED2K URIs provide a standardized way to reference files, servers, and node lists in the ed2k network.

### 2.1 URI Format

```
ed2k://|<type>|<parameters>|/
```

### 2.2 File Links

```
ed2k://|file|<filename>|<size>|<hash>|[<optional_params>]|/
```

**Parameters:**
- `filename`: URL-encoded filename
- `size`: File size in bytes (decimal)
- `hash`: 32-character hexadecimal MD4 hash
- `optional_params`: Additional metadata (sources, etc.)

**Example:**
```
ed2k://|file|example.mp3|5326710|398AF6A8B5EC0DAE0DC01B1D9A10C25C|/
```

### 2.3 Server Links

```
ed2k://|server|<hostname_or_ip>|<port>|/
```

**Example:**
```
ed2k://|server|176.103.48.36|4184|/
```

### 2.4 Server List Links

```
ed2k://|serverlist|<url>|/
```

### 2.5 Nodes List Links

```
ed2k://|nodeslist|<url>|/
```

### 2.6 Golang Implementation

```go
package protocol

import (
    "fmt"
    "net/url"
    "strconv"
    "strings"
)

type LinkType int

const (
    LinkTypeFile LinkType = iota
    LinkTypeServer
    LinkTypeServerList
    LinkTypeNodesList
)

type EMuleLink struct {
    Type     LinkType
    Filename string
    Size     int64
    Hash     Hash
    Host     string
    Port     uint16
    URL      string
}

func ParseEMuleLink(uri string) (*EMuleLink, error) {
    // URL decode
    decoded, err := url.QueryUnescape(uri)
    if err != nil {
        return nil, fmt.Errorf("failed to decode URI: %v", err)
    }
    
    // Split by pipes
    parts := strings.Split(decoded, "|")
    if len(parts) < 3 || parts[0] != "ed2k://" || parts[len(parts)-1] != "/" {
        return nil, fmt.Errorf("malformed ed2k URI")
    }
    
    link := &EMuleLink{}
    
    switch parts[1] {
    case "file":
        if len(parts) < 6 {
            return nil, fmt.Errorf("file link requires at least 5 parts")
        }
        link.Type = LinkTypeFile
        link.Filename, _ = url.QueryUnescape(parts[2])
        link.Size, err = strconv.ParseInt(parts[3], 10, 64)
        if err != nil {
            return nil, fmt.Errorf("invalid file size: %v", err)
        }
        link.Hash, err = HashFromString(parts[4])
        if err != nil {
            return nil, fmt.Errorf("invalid hash: %v", err)
        }
        
    case "server":
        if len(parts) != 5 {
            return nil, fmt.Errorf("server link requires exactly 4 parts")
        }
        link.Type = LinkTypeServer
        link.Host = parts[2]
        port, err := strconv.ParseUint(parts[3], 10, 16)
        if err != nil {
            return nil, fmt.Errorf("invalid port: %v", err)
        }
        link.Port = uint16(port)
        
    case "serverlist":
        if len(parts) != 4 {
            return nil, fmt.Errorf("serverlist link requires exactly 3 parts")
        }
        link.Type = LinkTypeServerList
        link.URL = parts[2]
        
    case "nodeslist":
        if len(parts) != 4 {
            return nil, fmt.Errorf("nodeslist link requires exactly 3 parts")
        }
        link.Type = LinkTypeNodesList
        link.URL = parts[2]
        
    default:
        return nil, fmt.Errorf("unknown link type: %s", parts[1])
    }
    
    return link, nil
}

func (l *EMuleLink) String() string {
    switch l.Type {
    case LinkTypeFile:
        encoded := url.QueryEscape(l.Filename)
        return fmt.Sprintf("ed2k://|file|%s|%d|%s|/", 
            encoded, l.Size, strings.ToUpper(l.Hash.String()))
    case LinkTypeServer:
        return fmt.Sprintf("ed2k://|server|%s|%d|/", l.Host, l.Port)
    case LinkTypeServerList:
        return fmt.Sprintf("ed2k://|serverlist|%s|/", l.URL)
    case LinkTypeNodesList:
        return fmt.Sprintf("ed2k://|nodeslist|%s|/", l.URL)
    }
    return ""
}
```

---

## 3. Protocol Fundamentals

### 3.1 Network Protocols

ED2K uses two transport protocols:
- **TCP**: Client-server and client-client communication
- **UDP**: Kademlia DHT and server queries

### 3.2 Packet Structure

All ed2k packets follow a common header format:

```
+----------+----------+----------+----------+----------+----------+
| Protocol |          Size           | Packet |    Data...      |
|  (1B)    |         (4B)            |  (1B)  |                 |
+----------+----------+----------+----------+----------+----------+
```

**Protocol Values:**
- `0xE3`: Standard eDonkey protocol
- `0xC5`: eMule extended protocol
- `0xD4`: Compressed protocol

### 3.3 Data Types

#### 3.3.1 Hash (16 bytes)
MD4 hash used for file and chunk identification.

```go
type Hash [16]byte

func (h Hash) String() string {
    return hex.EncodeToString(h[:])
}

func HashFromString(s string) (Hash, error) {
    var h Hash
    bytes, err := hex.DecodeString(s)
    if err != nil || len(bytes) != 16 {
        return h, fmt.Errorf("invalid hash format")
    }
    copy(h[:], bytes)
    return h, nil
}
```

#### 3.3.2 Endpoint (6 bytes)
Network endpoint (IPv4 address + port).

```go
type Endpoint struct {
    IP   uint32 // IPv4 address in network byte order
    Port uint16 // Port in network byte order
}

func (e *Endpoint) Serialize(buf []byte) {
    binary.LittleEndian.PutUint32(buf[0:4], e.IP)
    binary.LittleEndian.PutUint16(buf[4:6], e.Port)
}

func (e *Endpoint) Deserialize(buf []byte) error {
    if len(buf) < 6 {
        return fmt.Errorf("insufficient data for endpoint")
    }
    e.IP = binary.LittleEndian.Uint32(buf[0:4])
    e.Port = binary.LittleEndian.Uint16(buf[4:6])
    return nil
}

func (e *Endpoint) String() string {
    ip := net.IPv4(byte(e.IP), byte(e.IP>>8), byte(e.IP>>16), byte(e.IP>>24))
    return fmt.Sprintf("%s:%d", ip.String(), e.Port)
}
```

### 3.4 Packet Header Implementation

```go
type PacketHeader struct {
    Protocol byte
    Size     uint32
    Packet   byte
}

const (
    OpUndefined     = 0x00
    OpEDonkeyHeader = 0xE3
    OpEDonkeyProt   = 0xE3
    OpPackedProt    = 0xD4
    OpEMuleProt     = 0xC5
    HeaderSize      = 6
)

func (h *PacketHeader) Serialize() []byte {
    buf := make([]byte, HeaderSize)
    buf[0] = h.Protocol
    binary.LittleEndian.PutUint32(buf[1:5], h.Size)
    buf[5] = h.Packet
    return buf
}

func (h *PacketHeader) Deserialize(buf []byte) error {
    if len(buf) < HeaderSize {
        return fmt.Errorf("insufficient data for packet header")
    }
    h.Protocol = buf[0]
    h.Size = binary.LittleEndian.Uint32(buf[1:5])
    h.Packet = buf[5]
    return nil
}

func (h *PacketHeader) PayloadSize() int {
    return int(h.Size) - 1 // Size includes packet type byte
}
```

---

## 4. Client-Server Communication

### 4.1 Connection Establishment

Clients connect to servers using TCP on the standard port (typically 4661).

### 4.2 Client User Hash Generation Specification

**Client User Hash** is a unique 16-byte identifier for each ed2k client, used for network identification and tracking. The generation follows specific rules based on the Java implementation.

#### 4.2.1 User Agent Hash Constants

Each client software has a specific user agent hash for identification:

```go
// Predefined user agent hash constants
const (
    // Official eDonkey client
    UserAgentEDonkey = "31D6CFE0D16AE931B73C59D7E0C089C0"
    
    // eMule client (most common)
    UserAgentEmule = "31D6CFE0D10EE931B73C59D7E0C06FC0"
    
    // libed2k library
    UserAgentLibed2k = "31D6CFE0D14CE931B73C59D7E0C04BC0"
)

var (
    HashTerminal = HashFromString("31D6CFE0D16AE931B73C59D7E0C089C0")
    HashLibed2k  = HashFromString("31D6CFE0D14CE931B73C59D7E0C04BC0") 
    HashEmule    = HashFromString("31D6CFE0D10EE931B73C59D7E0C06FC0")
)

type ClientSettings struct {
    UserAgent Hash   // Client user agent hash
    ModName   string // Module name
    Version   int    // Client version
}

func NewClientSettings() *ClientSettings {
    return &ClientSettings{
        UserAgent: HashFromString(UserAgentEmule), // Default to eMule identification
        ModName:   "ged2k",
        Version:   0x3c,
    }
}
```

#### 4.2.2 Random Client Hash Generation

For normal client identification, generate a random 16-byte hash following these rules:

```go
import (
    "crypto/rand"
    "crypto/md4"
    "time"
)

// Generate random client user hash
func GenerateClientUserHash(isEmuleCompatible bool) Hash {
    // Generate 16 bytes of random data
    randomBytes := make([]byte, 16)
    _, err := rand.Read(randomBytes)
    if err != nil {
        // Fallback to timestamp-based generation
        timestamp := time.Now().UnixNano()
        for i := 0; i < 16; i += 8 {
            binary.LittleEndian.PutUint64(randomBytes[i:], uint64(timestamp+int64(i)))
        }
    }
    
    // Set eMule compatibility markers if requested
    if isEmuleCompatible {
        randomBytes[5] = 14   // 6th byte = 14 (eMule identification)
        randomBytes[14] = 111 // 15th byte = 111 (eMule identification)
    }
    
    var hash Hash
    copy(hash[:], randomBytes)
    return hash
}

// Generate deterministic user hash from seed (MAC address, etc.)
func GenerateDeterministicUserHash(seed string, isEmuleCompatible bool) Hash {
    hasher := md4.New()
    hasher.Write([]byte(seed))
    hasher.Write([]byte(time.Now().Format("2006-01-02"))) // Daily variation
    
    hashBytes := hasher.Sum(nil)
    
    if isEmuleCompatible {
        hashBytes[5] = 14
        hashBytes[14] = 111
    }
    
    var hash Hash
    copy(hash[:], hashBytes)
    return hash
}
```

**Design Rationale:** The user hash serves multiple purposes:
1. **Server Login**: Identifies the client to servers
2. **Peer Connections**: Used in peer-to-peer handshakes
3. **Transfer Tracking**: Servers track upload/download statistics per client
4. **Reconnection**: Allows servers to recognize returning clients

The eMule compatibility bytes (5th=14, 14th=111) ensure compatibility with the existing eMule network and proper client recognition.

#### 4.2.3 Hash Usage in Protocol

The client user hash is used in several protocol operations:

```go
// Login request with generated user hash
func LoginToServer(conn net.Conn, settings *ClientSettings) error {
    userHash := settings.UserAgent
    if userHash == (Hash{}) {
        userHash = GenerateClientUserHash(true) // Generate eMule-compatible hash
    }
    
    loginReq := &LoginRequest{
        Hash: userHash,
        Port: uint16(settings.ListenPort),
        Tags: []Tag{
            NewTag(TagTypeString, "name", settings.ModName),
            NewTag(TagTypeUInt32, "version", uint32(settings.Version)),
        },
    }
    
    return sendLoginRequest(conn, loginReq)
}

// Peer handshake with user hash
func CreateHelloPacket(userHash Hash, nick string, port uint16) *HelloPacket {
    return &HelloPacket{
        UserHash: userHash,
        Port:     port,
        Properties: []Tag{
            NewTag(TagTypeString, "", nick),
            NewTag(TagTypeUInt32, "", makeEmuleVersion(0, 60, 0)),
        },
    }
}
```

### 4.3 Login Sequence

1. **Client → Server: Login Request (0x01)**
2. **Server → Client: ID Change (0x40)** 
3. **Server → Client: Server Message (0x38)** (optional)

#### 4.3.1 Login Request Packet

```
+----------+----------+----------+----------+----------+----------+
| Hash (16 bytes)     | IP (4B)  | Port(2B) | Tags...           |
+----------+----------+----------+----------+----------+----------+
```

**Design Rationale:** The login request includes the client's user hash (persistent identifier), current IP/port for callback connections, and capability tags. This allows the server to assign an appropriate client ID and understand client capabilities.

```go
type LoginRequest struct {
    Hash Hash
    IP   uint32
    Port uint16
    Tags []Tag
}

const OpLoginRequest = 0x01

func (l *LoginRequest) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    
    // Write hash
    buf.Write(l.Hash[:])
    
    // Write IP and port
    binary.Write(buf, binary.LittleEndian, l.IP)
    binary.Write(buf, binary.LittleEndian, l.Port)
    
    // Write tags
    binary.Write(buf, binary.LittleEndian, uint32(len(l.Tags)))
    for _, tag := range l.Tags {
        tagData := tag.Serialize()
        buf.Write(tagData)
    }
    
    return buf.Bytes()
}

// Standard client tags
func CreateStandardTags() []Tag {
    return []Tag{
        {Type: TagTypeString, Name: "name", Value: "ged2k-go"},
        {Type: TagTypeUint32, Name: "version", Value: uint32(0x3c)},
        {Type: TagTypeUint32, Name: "flags", Value: uint32(0x0d)},
    }
}
```

#### 4.1.3 ID Change Response

```
+----------+
| ID (4B)  |
+----------+
```

The server assigns a Client ID to the connecting client.

```go
type IdChange struct {
    ClientID uint32
}

const OpIdChange = 0x40

func (i *IdChange) Deserialize(buf []byte) error {
    if len(buf) < 4 {
        return fmt.Errorf("insufficient data for ID change")
    }
    i.ClientID = binary.LittleEndian.Uint32(buf[0:4])
    return nil
}
```

### 4.2 File Search Protocol

#### 4.2.1 Search Request (0x16)

```
+----------+----------+----------+----------+----------+
| Search Type(1B) | Query String... | File Type | Size Constraints |
+----------+----------+----------+----------+----------+
```

**Search Types:**
- `0x00`: Generic search
- `0x01`: Audio files
- `0x02`: Video files
- `0x03`: Images
- `0x04`: Documents
- `0x05`: Programs

```go
type SearchRequest struct {
    SearchType uint8
    Query      string
    FileType   uint8
    MinSize    uint64
    MaxSize    uint64
    Extension  string
}

const OpSearchRequest = 0x16

func (s *SearchRequest) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    
    // Write search type
    buf.WriteByte(s.SearchType)
    
    // Write query string
    queryBytes := []byte(s.Query)
    binary.Write(buf, binary.LittleEndian, uint16(len(queryBytes)))
    buf.Write(queryBytes)
    
    // Write file type constraints
    if s.FileType != 0 {
        buf.WriteByte(TagTypeUint8)
        buf.WriteByte(1) // Name length (tag ID)
        buf.WriteByte(0x01) // File type tag ID
        buf.WriteByte(s.FileType)
    }
    
    // Write size constraints
    if s.MinSize > 0 {
        buf.WriteByte(TagTypeUint64)
        buf.WriteByte(1)
        buf.WriteByte(0x02) // Min size tag ID
        binary.Write(buf, binary.LittleEndian, s.MinSize)
    }
    
    if s.MaxSize > 0 {
        buf.WriteByte(TagTypeUint64)
        buf.WriteByte(1)
        buf.WriteByte(0x03) // Max size tag ID
        binary.Write(buf, binary.LittleEndian, s.MaxSize)
    }
    
    return buf.Bytes()
}
```

#### 4.2.2 Search Result (0x17)

```
+----------+----------+----------+----------+----------+
| Count(4B)| File Hash (16B)  | IP(4B) | Port(2B) | Tags... |
+----------+----------+----------+----------+----------+
```

**Design Rationale:** Search results include file metadata (name, size, type) and peer information (IP/port). The tag system allows flexible metadata without breaking protocol compatibility.

```go
type SearchResult struct {
    Hash Hash
    IP   uint32
    Port uint16
    Tags []Tag
}

const OpSearchResult = 0x17

func ParseSearchResults(buf []byte) ([]SearchResult, error) {
    if len(buf) < 4 {
        return nil, fmt.Errorf("insufficient data")
    }
    
    count := binary.LittleEndian.Uint32(buf[0:4])
    results := make([]SearchResult, 0, count)
    offset := 4
    
    for i := uint32(0); i < count && offset < len(buf); i++ {
        if offset+16+4+2 > len(buf) {
            break
        }
        
        result := SearchResult{}
        copy(result.Hash[:], buf[offset:offset+16])
        offset += 16
        
        result.IP = binary.LittleEndian.Uint32(buf[offset : offset+4])
        offset += 4
        result.Port = binary.LittleEndian.Uint16(buf[offset : offset+2])
        offset += 2
        
        // Parse tags
        if offset+4 <= len(buf) {
            tagCount := binary.LittleEndian.Uint32(buf[offset : offset+4])
            offset += 4
            
            for j := uint32(0); j < tagCount && offset < len(buf); j++ {
                tag, consumed, err := ParseTag(buf[offset:])
                if err != nil {
                    break
                }
                result.Tags = append(result.Tags, tag)
                offset += consumed
            }
        }
        
        results = append(results, result)
    }
    
    return results, nil
}
```

### 4.3 File Source Requests

#### 4.3.1 Get File Sources (0x19)

Request peer list for a specific file hash.

```go
type GetFileSources struct {
    Hash Hash
}

const OpGetFileSources = 0x19

func (g *GetFileSources) Serialize() []byte {
    return g.Hash[:]
}
```

#### 4.3.2 Found File Sources (0x1A)

```
+----------+----------+----------+----------+
| Hash (16 bytes)     | Count(1B)| Endpoints... |
+----------+----------+----------+----------+
```

```go
type FoundFileSources struct {
    Hash    Hash
    Sources []Endpoint
}

const OpFoundFileSources = 0x1A

func (f *FoundFileSources) Deserialize(buf []byte) error {
    if len(buf) < 17 {
        return fmt.Errorf("insufficient data")
    }
    
    copy(f.Hash[:], buf[0:16])
    count := buf[16]
    offset := 17
    
    f.Sources = make([]Endpoint, 0, count)
    for i := byte(0); i < count && offset+6 <= len(buf); i++ {
        var endpoint Endpoint
        err := endpoint.Deserialize(buf[offset : offset+6])
        if err != nil {
            return err
        }
        f.Sources = append(f.Sources, endpoint)
        offset += 6
    }
    
    return nil
}
```

---

## 5. Client-Client Communication

### 5.1 Connection Establishment

Client-to-client connections use TCP on dynamically assigned ports.

#### 5.1.1 Handshake Sequence

1. **Initiator → Target: Hello (0x01)**
2. **Target → Initiator: Hello Answer (0x02)**

#### 5.1.2 Hello Packet

```
+----------+----------+----------+----------+----------+
| Hash (16 bytes)     | ID (4B)  | Port(2B) | Tags...  |
+----------+----------+----------+----------+----------+
```

**Design Rationale:** The Hello packet establishes peer identity and capabilities. The hash identifies the client persistently, while tags communicate supported features and client software version.

```go
type Hello struct {
    Hash Hash
    ID   uint32
    Port uint16
    Tags []Tag
}

const OpHello = 0x01

func (h *Hello) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    
    buf.Write(h.Hash[:])
    binary.Write(buf, binary.LittleEndian, h.ID)
    binary.Write(buf, binary.LittleEndian, h.Port)
    
    binary.Write(buf, binary.LittleEndian, uint32(len(h.Tags)))
    for _, tag := range h.Tags {
        buf.Write(tag.Serialize())
    }
    
    return buf.Bytes()
}

func (h *Hello) Deserialize(buf []byte) error {
    if len(buf) < 22 { // Hash + ID + Port + TagCount
        return fmt.Errorf("insufficient data")
    }
    
    offset := 0
    copy(h.Hash[:], buf[offset:offset+16])
    offset += 16
    
    h.ID = binary.LittleEndian.Uint32(buf[offset : offset+4])
    offset += 4
    h.Port = binary.LittleEndian.Uint16(buf[offset : offset+2])
    offset += 2
    
    tagCount := binary.LittleEndian.Uint32(buf[offset : offset+4])
    offset += 4
    
    h.Tags = make([]Tag, 0, tagCount)
    for i := uint32(0); i < tagCount && offset < len(buf); i++ {
        tag, consumed, err := ParseTag(buf[offset:])
        if err != nil {
            return err
        }
        h.Tags = append(h.Tags, tag)
        offset += consumed
    }
    
    return nil
}
```

### 5.2 File Transfer Protocol

#### 5.2.1 File Request (0x58)

Request to download a file.

```
+----------+----------+----------+----------+
| Hash (16 bytes)     | Part Mask...       |
+----------+----------+----------+----------+
```

**Design Rationale:** File requests include a bit mask indicating which 9.28MB chunks are needed. This enables efficient resume of partial downloads.

```go
type FileRequest struct {
    Hash     Hash
    PartMask []byte // Bit mask for 9.28MB chunks
}

const OpFileRequest = 0x58

func (f *FileRequest) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    buf.Write(f.Hash[:])
    
    // Write part mask size and data
    if len(f.PartMask) > 0 {
        binary.Write(buf, binary.LittleEndian, uint16(len(f.PartMask)))
        buf.Write(f.PartMask)
    } else {
        binary.Write(buf, binary.LittleEndian, uint16(0))
    }
    
    return buf.Bytes()
}
```

#### 5.2.2 Request Parts (0x47)

Request specific data blocks.

```
+----------+----------+----------+----------+----------+
| Hash (16 bytes)     | Start Offset(8B) | End Offset(8B) |
+----------+----------+----------+----------+----------+
```

**Modern 64-bit variant for large files:**

```go
type RequestParts64 struct {
    Hash        Hash
    StartOffset uint64
    EndOffset   uint64
}

const OpRequestParts64 = 0x47

func (r *RequestParts64) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    buf.Write(r.Hash[:])
    binary.Write(buf, binary.LittleEndian, r.StartOffset)
    binary.Write(buf, binary.LittleEndian, r.EndOffset)
    return buf.Bytes()
}

func (r *RequestParts64) Deserialize(buf []byte) error {
    if len(buf) < 32 { // 16 + 8 + 8
        return fmt.Errorf("insufficient data")
    }
    
    copy(r.Hash[:], buf[0:16])
    r.StartOffset = binary.LittleEndian.Uint64(buf[16:24])
    r.EndOffset = binary.LittleEndian.Uint64(buf[24:32])
    
    return nil
}
```

#### 5.2.3 Sending Part (0x46)

Send requested data block.

```
+----------+----------+----------+----------+----------+
| Hash (16 bytes)     | Start Offset(8B) | End Offset(8B) | Data... |
+----------+----------+----------+----------+----------+
```

```go
type SendingPart64 struct {
    Hash        Hash
    StartOffset uint64
    EndOffset   uint64
    Data        []byte
}

const OpSendingPart64 = 0x46

func (s *SendingPart64) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    buf.Write(s.Hash[:])
    binary.Write(buf, binary.LittleEndian, s.StartOffset)
    binary.Write(buf, binary.LittleEndian, s.EndOffset)
    buf.Write(s.Data)
    return buf.Bytes()
}

func (s *SendingPart64) Deserialize(buf []byte) error {
    if len(buf) < 32 {
        return fmt.Errorf("insufficient data")
    }
    
    copy(s.Hash[:], buf[0:16])
    s.StartOffset = binary.LittleEndian.Uint64(buf[16:24])
    s.EndOffset = binary.LittleEndian.Uint64(buf[24:32])
    s.Data = make([]byte, len(buf)-32)
    copy(s.Data, buf[32:])
    
    return nil
}
```

### 5.3 Hash Set Exchange

#### 5.3.1 Hash Set Request (0x51)

Request chunk hashes for integrity verification.

```go
type HashSetRequest struct {
    Hash Hash
}

const OpHashSetRequest = 0x51

func (h *HashSetRequest) Serialize() []byte {
    return h.Hash[:]
}
```

#### 5.3.2 Hash Set Answer (0x52)

Provide chunk hash list.

```
+----------+----------+----------+----------+
| Hash (16 bytes)     | Count(2B)| Hashes... |
+----------+----------+----------+----------+
```

**Design Rationale:** Each 9.28MB chunk has its own MD4 hash for integrity verification. This allows detection of corruption in specific chunks without re-downloading the entire file.

```go
type HashSetAnswer struct {
    Hash   Hash
    Hashes []Hash
}

const OpHashSetAnswer = 0x52

func (h *HashSetAnswer) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    buf.Write(h.Hash[:])
    
    binary.Write(buf, binary.LittleEndian, uint16(len(h.Hashes)))
    for _, hash := range h.Hashes {
        buf.Write(hash[:])
    }
    
    return buf.Bytes()
}
```

---

## 6. Kademlia Network Protocol

The Kademlia DHT provides decentralized peer discovery and file indexing.

### 6.1 Node Identification

Each Kademlia node has a 128-bit identifier (KadID) derived from its IP address and a random component.

```go
type KadID [16]byte

func (k KadID) Distance(other KadID) KadID {
    var result KadID
    for i := 0; i < 16; i++ {
        result[i] = k[i] ^ other[i]
    }
    return result
}

func (k KadID) String() string {
    return hex.EncodeToString(k[:])
}
```

### 6.2 Packet Structure

Kademlia packets use a different header format:

```
+----------+----------+----------+----------+
| Opcode(1B)| Peer ID (16B)     | Data...  |
+----------+----------+----------+----------+
```

### 6.3 Bootstrap Protocol

#### 6.3.1 Bootstrap Request (0x01)

```go
type Kad2BootstrapReq struct{}

const KadOpBootstrapReq = 0x01

func (k *Kad2BootstrapReq) Serialize() []byte {
    return []byte{} // No payload
}
```

#### 6.3.2 Bootstrap Response (0x09)

```
+----------+----------+----------+----------+
| Count(2B)| Contact 1 (25B) | Contact 2... |
+----------+----------+----------+----------+
```

Each contact contains:
- KadID (16 bytes)
- IP address (4 bytes)
- TCP port (2 bytes)
- UDP port (2 bytes)
- Version (1 byte)

```go
type KadContact struct {
    ID       KadID
    IP       uint32
    TCPPort  uint16
    UDPPort  uint16
    Version  uint8
}

type Kad2BootstrapRes struct {
    Contacts []KadContact
}

const KadOpBootstrapRes = 0x09

func (k *Kad2BootstrapRes) Deserialize(buf []byte) error {
    if len(buf) < 2 {
        return fmt.Errorf("insufficient data")
    }
    
    count := binary.LittleEndian.Uint16(buf[0:2])
    offset := 2
    
    k.Contacts = make([]KadContact, 0, count)
    
    for i := uint16(0); i < count && offset+25 <= len(buf); i++ {
        contact := KadContact{}
        
        copy(contact.ID[:], buf[offset:offset+16])
        offset += 16
        
        contact.IP = binary.LittleEndian.Uint32(buf[offset : offset+4])
        offset += 4
        contact.TCPPort = binary.LittleEndian.Uint16(buf[offset : offset+2])
        offset += 2
        contact.UDPPort = binary.LittleEndian.Uint16(buf[offset : offset+2])
        offset += 2
        contact.Version = buf[offset]
        offset++
        
        k.Contacts = append(k.Contacts, contact)
    }
    
    return nil
}
```

### 6.4 Node Lookup Protocol

#### 6.4.1 Hello Request (0x11)

```go
type Kad2HelloReq struct {
    ReceiverID KadID
    TCPPort    uint16
    Version    uint8
}

const KadOpHelloReq = 0x11
```

#### 6.4.2 Search Requests

**Search Sources Request (0x52):**
Find peers sharing a specific file.

```
+----------+----------+----------+----------+
| Target Hash (16B)   | Start ID (16B)     |
+----------+----------+----------+----------+
```

```go
type Kad2SearchSourcesReq struct {
    FileHash Hash
    StartID  KadID
}

const KadOpSearchSourcesReq = 0x52

func (k *Kad2SearchSourcesReq) Serialize() []byte {
    buf := make([]byte, 32)
    copy(buf[0:16], k.FileHash[:])
    copy(buf[16:32], k.StartID[:])
    return buf
}
```

---

## 7. High ID and Low ID System

The ED2K protocol uses a client ID system to determine connectivity capabilities.

### 7.1 ID Assignment

**High ID (> 16777216):** 
- Client has full internet connectivity
- Can accept incoming connections
- Assigned actual IP address as ID

**Low ID (< 16777216):**
- Client behind NAT/firewall  
- Cannot accept direct connections
- Assigned arbitrary low number

**Design Rationale:** This system was designed before modern NAT traversal techniques. High ID clients can serve as connection brokers for Low ID clients, enabling the network to function despite NAT/firewall restrictions.

### 7.2 ID Determination Logic

```go
func DetermineClientID(ip net.IP, canAcceptConnections bool) uint32 {
    if canAcceptConnections && !isPrivateIP(ip) {
        // High ID: use actual IP address
        ipv4 := ip.To4()
        if ipv4 != nil {
            return binary.LittleEndian.Uint32(ipv4)
        }
    }
    
    // Low ID: assign random number < 16777216
    return rand.Uint32() % 16777216
}

func isPrivateIP(ip net.IP) bool {
    ipv4 := ip.To4()
    if ipv4 == nil {
        return false
    }
    
    // Check private ranges
    return ipv4[0] == 10 ||
           (ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31) ||
           (ipv4[0] == 192 && ipv4[1] == 168)
}

func IsHighID(id uint32) bool {
    return id > 16777216
}

func IsLowID(id uint32) bool {
    return id <= 16777216 && id != 0
}
```

### 7.3 Connection Handling

```go
type ConnectionManager struct {
    clientID    uint32
    isHighID    bool
    callbackReq func(targetID uint32, endpoint Endpoint) error
}

func (cm *ConnectionManager) ConnectToPeer(targetID uint32, endpoint Endpoint) error {
    if cm.isHighID {
        // High ID can connect directly
        return cm.directConnect(endpoint)
    } else {
        // Low ID must request callback
        return cm.callbackReq(targetID, endpoint)
    }
}

func (cm *ConnectionManager) AcceptConnection(conn net.Conn) error {
    if !cm.isHighID {
        // Low ID clients should not accept connections
        conn.Close()
        return fmt.Errorf("low ID client cannot accept connections")
    }
    
    return cm.handleIncomingConnection(conn)
}
```

---

## 8. Server.met File Specification

The server.met file contains a list of known ED2K servers with their metadata.

### 8.1 File Format

```
+----------+----------+----------+----------+
| Version  | Count(4B)| Server 1 | Server 2...
|  (1B)    |          | Entry    | Entry    |
+----------+----------+----------+----------+
```

### 8.2 Server Entry Format

```
+----------+----------+----------+----------+----------+
| IP (4B)  | Port(2B) | Tag Count| Tags...           |
+----------+----------+----------+----------+----------+
```

**Design Rationale:** The tag-based format allows flexible server metadata without breaking compatibility. New server properties can be added as new tag types.

### 8.3 Server Tags

Common server tags:
- `0x01`: Server name (string)
- `0x0B`: Description (string)
- `0x0C`: Ping (uint32)
- `0x87`: Max users (uint32)
- `0x88`: Soft files limit (uint32)
- `0x89`: Hard files limit (uint32)
- `0x91`: Version (string)
- `0x92`: UDP flags (uint32)

### 8.4 Implementation

```go
type Server struct {
    IP          uint32
    Port        uint16
    Name        string
    Description string
    Ping        uint32
    MaxUsers    uint32
    Files       uint32
    Version     string
    UDPFlags    uint32
    Tags        []Tag
}

type ServerMet struct {
    Version byte
    Servers []Server
}

func ParseServerMet(data []byte) (*ServerMet, error) {
    if len(data) < 5 {
        return nil, fmt.Errorf("file too short")
    }
    
    serverMet := &ServerMet{}
    serverMet.Version = data[0]
    
    count := binary.LittleEndian.Uint32(data[1:5])
    offset := 5
    
    for i := uint32(0); i < count && offset < len(data); i++ {
        if offset+6 > len(data) {
            break
        }
        
        server := Server{}
        server.IP = binary.LittleEndian.Uint32(data[offset : offset+4])
        offset += 4
        server.Port = binary.LittleEndian.Uint16(data[offset : offset+2])
        offset += 2
        
        if offset+4 > len(data) {
            break
        }
        
        tagCount := binary.LittleEndian.Uint32(data[offset : offset+4])
        offset += 4
        
        // Parse tags
        for j := uint32(0); j < tagCount && offset < len(data); j++ {
            tag, consumed, err := ParseServerTag(data[offset:])
            if err != nil {
                break
            }
            
            // Apply tag to server fields
            switch tag.ID {
            case 0x01: // Server name
                if str, ok := tag.Value.(string); ok {
                    server.Name = str
                }
            case 0x0B: // Description
                if str, ok := tag.Value.(string); ok {
                    server.Description = str
                }
            case 0x0C: // Ping
                if val, ok := tag.Value.(uint32); ok {
                    server.Ping = val
                }
            case 0x87: // Max users
                if val, ok := tag.Value.(uint32); ok {
                    server.MaxUsers = val
                }
            case 0x91: // Version
                if str, ok := tag.Value.(string); ok {
                    server.Version = str
                }
            case 0x92: // UDP flags
                if val, ok := tag.Value.(uint32); ok {
                    server.UDPFlags = val
                }
            }
            
            server.Tags = append(server.Tags, tag)
            offset += consumed
        }
        
        serverMet.Servers = append(serverMet.Servers, server)
    }
    
    return serverMet, nil
}

type ServerTag struct {
    Type  byte
    ID    byte
    Name  string
    Value interface{}
}

func ParseServerTag(data []byte) (ServerTag, int, error) {
    if len(data) < 2 {
        return ServerTag{}, 0, fmt.Errorf("insufficient data")
    }
    
    tag := ServerTag{}
    tag.Type = data[0]
    nameLen := data[1]
    offset := 2
    
    if nameLen == 1 {
        // Tag ID follows
        if offset >= len(data) {
            return tag, 0, fmt.Errorf("insufficient data for tag ID")
        }
        tag.ID = data[offset]
        offset++
    } else {
        // Tag name follows
        if offset+int(nameLen) > len(data) {
            return tag, 0, fmt.Errorf("insufficient data for tag name")
        }
        tag.Name = string(data[offset : offset+int(nameLen)])
        offset += int(nameLen)
    }
    
    // Parse value based on type
    switch tag.Type {
    case 0x02: // String
        if offset+2 > len(data) {
            return tag, 0, fmt.Errorf("insufficient data for string length")
        }
        strLen := binary.LittleEndian.Uint16(data[offset : offset+2])
        offset += 2
        
        if offset+int(strLen) > len(data) {
            return tag, 0, fmt.Errorf("insufficient data for string value")
        }
        tag.Value = string(data[offset : offset+int(strLen)])
        offset += int(strLen)
        
    case 0x03: // UInt32
        if offset+4 > len(data) {
            return tag, 0, fmt.Errorf("insufficient data for uint32")
        }
        tag.Value = binary.LittleEndian.Uint32(data[offset : offset+4])
        offset += 4
        
    case 0x08: // UInt16
        if offset+2 > len(data) {
            return tag, 0, fmt.Errorf("insufficient data for uint16")
        }
        tag.Value = binary.LittleEndian.Uint16(data[offset : offset+2])
        offset += 2
        
    case 0x09: // UInt8
        if offset >= len(data) {
            return tag, 0, fmt.Errorf("insufficient data for uint8")
        }
        tag.Value = data[offset]
        offset++
        
    default:
        return tag, 0, fmt.Errorf("unknown tag type: %d", tag.Type)
    }
    
    return tag, offset, nil
}
```

---

## 9. Nodes.dat File Specification

The nodes.dat file contains Kademlia bootstrap nodes for DHT initialization.

### 9.1 File Format

```
+----------+----------+----------+----------+
| Version  | Count(4B)| Node 1   | Node 2...
|  (4B)    |          | Entry    | Entry    |
+----------+----------+----------+----------+
```

### 9.2 Node Entry Format

```
+----------+----------+----------+----------+----------+----------+
| KadID (16 bytes)     | IP (4B)  | UDP Port | TCP Port | Ver(1B)|
|                      |          |  (2B)    |  (2B)    |        |
+----------+----------+----------+----------+----------+----------+
```

**Design Rationale:** Nodes.dat provides a bootstrap mechanism for new clients to join the Kademlia network. The file format is compact and includes sufficient information to establish initial DHT connections.

### 9.3 Implementation

```go
type KadNode struct {
    ID      KadID
    IP      uint32
    UDPPort uint16
    TCPPort uint16
    Version byte
}

type NodesDat struct {
    Version uint32
    Nodes   []KadNode
}

func ParseNodesDat(data []byte) (*NodesDat, error) {
    if len(data) < 8 {
        return nil, fmt.Errorf("file too short")
    }
    
    nodesDat := &NodesDat{}
    nodesDat.Version = binary.LittleEndian.Uint32(data[0:4])
    count := binary.LittleEndian.Uint32(data[4:8])
    
    offset := 8
    nodesDat.Nodes = make([]KadNode, 0, count)
    
    for i := uint32(0); i < count && offset+25 <= len(data); i++ {
        node := KadNode{}
        
        // Parse KadID (16 bytes)
        copy(node.ID[:], data[offset:offset+16])
        offset += 16
        
        // Parse IP address (4 bytes)
        node.IP = binary.LittleEndian.Uint32(data[offset : offset+4])
        offset += 4
        
        // Parse UDP port (2 bytes)
        node.UDPPort = binary.LittleEndian.Uint16(data[offset : offset+2])
        offset += 2
        
        // Parse TCP port (2 bytes)
        node.TCPPort = binary.LittleEndian.Uint16(data[offset : offset+2])
        offset += 2
        
        // Parse version (1 byte)
        node.Version = data[offset]
        offset++
        
        nodesDat.Nodes = append(nodesDat.Nodes, node)
    }
    
    return nodesDat, nil
}

func (n *NodesDat) Serialize() []byte {
    buf := bytes.NewBuffer(nil)
    
    binary.Write(buf, binary.LittleEndian, n.Version)
    binary.Write(buf, binary.LittleEndian, uint32(len(n.Nodes)))
    
    for _, node := range n.Nodes {
        buf.Write(node.ID[:])
        binary.Write(buf, binary.LittleEndian, node.IP)
        binary.Write(buf, binary.LittleEndian, node.UDPPort)
        binary.Write(buf, binary.LittleEndian, node.TCPPort)
        buf.WriteByte(node.Version)
    }
    
    return buf.Bytes()
}

// Helper function to convert IP to string
func (n *KadNode) IPString() string {
    ip := net.IPv4(byte(n.IP), byte(n.IP>>8), byte(n.IP>>16), byte(n.IP>>24))
    return ip.String()
}

// Bootstrap using nodes.dat
func Bootstrap(nodesDat *NodesDat, localKadID KadID) error {
    for _, node := range nodesDat.Nodes {
        // Skip nodes that are too close (avoid self)
        distance := localKadID.Distance(node.ID)
        if isZero(distance[:4]) { // Too close
            continue
        }
        
        // Send bootstrap request
        err := sendBootstrapRequest(node)
        if err != nil {
            log.Printf("Bootstrap failed for node %s: %v", node.IPString(), err)
            continue
        }
        
        log.Printf("Bootstrap request sent to %s:%d", node.IPString(), node.UDPPort)
    }
    
    return nil
}
```

---

## 10. Security Considerations

### 10.1 Hash Verification

All file chunks must be verified against their MD4 hashes to prevent corruption and malicious modification.

```go
func VerifyChunk(data []byte, expectedHash Hash) bool {
    h := md4.New()
    h.Write(data)
    actualHash := h.Sum(nil)
    
    return bytes.Equal(actualHash, expectedHash[:])
}
```

### 10.2 Connection Limits

Implement connection limits to prevent resource exhaustion:

```go
type ConnectionLimiter struct {
    maxConnections    int
    currentConnections int32
    perIPLimit        map[string]int
    mutex             sync.RWMutex
}

func (cl *ConnectionLimiter) AllowConnection(remoteAddr string) bool {
    cl.mutex.Lock()
    defer cl.mutex.Unlock()
    
    if atomic.LoadInt32(&cl.currentConnections) >= int32(cl.maxConnections) {
        return false
    }
    
    ip := strings.Split(remoteAddr, ":")[0]
    if cl.perIPLimit[ip] >= 5 { // Max 5 connections per IP
        return false
    }
    
    atomic.AddInt32(&cl.currentConnections, 1)
    cl.perIPLimit[ip]++
    return true
}
```

### 10.3 Packet Size Limits

Enforce maximum packet sizes to prevent buffer overflows:

```go
const MaxPacketSize = 1024 * 1024 // 1MB limit

func ReadPacket(conn net.Conn) ([]byte, error) {
    header := make([]byte, HeaderSize)
    _, err := io.ReadFull(conn, header)
    if err != nil {
        return nil, err
    }
    
    packetHeader := &PacketHeader{}
    err = packetHeader.Deserialize(header)
    if err != nil {
        return nil, err
    }
    
    if packetHeader.PayloadSize() > MaxPacketSize {
        return nil, fmt.Errorf("packet too large: %d bytes", packetHeader.PayloadSize())
    }
    
    payload := make([]byte, packetHeader.PayloadSize())
    _, err = io.ReadFull(conn, payload)
    return payload, err
}
```

---

## 11. Implementation Guidelines

### 11.1 Complete Client Implementation

```go
package main

import (
    "fmt"
    "log"
    "net"
    "time"
)

type ED2KClient struct {
    settings      *Settings
    session       *Session
    serverConn    *ServerConnection
    transfers     map[Hash]*Transfer
    peerConns     map[string]*PeerConnection
}

func NewED2KClient(settings *Settings) *ED2KClient {
    return &ED2KClient{
        settings:  settings,
        transfers: make(map[Hash]*Transfer),
        peerConns: make(map[string]*PeerConnection),
    }
}

func (c *ED2KClient) Start() error {
    // Initialize session
    c.session = NewSession(c.settings)
    
    // Load server list from server.met
    servers, err := c.loadServerList()
    if err != nil {
        log.Printf("Warning: Could not load server list: %v", err)
    }
    
    // Connect to servers
    for _, server := range servers {
        err = c.connectToServer(server)
        if err == nil {
            break // Connected to at least one server
        }
    }
    
    // Bootstrap Kademlia
    nodes, err := c.loadNodesList()
    if err == nil {
        c.bootstrapKademlia(nodes)
    }
    
    // Start listening for incoming connections
    return c.startListener()
}

func (c *ED2KClient) AddDownload(link string) error {
    emuleLink, err := ParseEMuleLink(link)
    if err != nil {
        return err
    }
    
    if emuleLink.Type != LinkTypeFile {
        return fmt.Errorf("not a file link")
    }
    
    transfer := NewTransfer(emuleLink.Hash, emuleLink.Filename, emuleLink.Size)
    c.transfers[emuleLink.Hash] = transfer
    
    // Request peers from server
    if c.serverConn != nil {
        c.serverConn.GetFileSources(emuleLink.Hash)
    }
    
    return nil
}

func (c *ED2KClient) Search(query string, fileType uint8) error {
    if c.serverConn == nil {
        return fmt.Errorf("not connected to server")
    }
    
    searchReq := &SearchRequest{
        SearchType: 0x00, // Generic search
        Query:      query,
        FileType:   fileType,
    }
    
    return c.serverConn.Search(searchReq)
}
```

### 11.2 Performance Optimization

#### 11.2.1 Connection Pooling

```go
type PeerConnectionPool struct {
    connections chan *PeerConnection
    factory     func() *PeerConnection
    maxSize     int
}

func (p *PeerConnectionPool) Get() *PeerConnection {
    select {
    case conn := <-p.connections:
        return conn
    default:
        return p.factory()
    }
}

func (p *PeerConnectionPool) Put(conn *PeerConnection) {
    select {
    case p.connections <- conn:
    default:
        conn.Close() // Pool is full, close connection
    }
}
```

#### 11.2.2 Concurrent Downloads

```go
func (t *Transfer) downloadConcurrently() {
    semaphore := make(chan struct{}, 10) // Limit to 10 concurrent downloads
    
    for _, chunk := range t.neededChunks {
        go func(c *Chunk) {
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            t.downloadChunk(c)
        }(chunk)
    }
}
```

### 11.3 Error Handling

```go
type ED2KError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

func (e *ED2KError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

const (
    ErrHashMismatch ErrorCode = iota
    ErrConnectionFailed
    ErrTimeout
    ErrInvalidPacket
)
```

---

## 12. Conclusion

This specification provides a complete reference for implementing the ED2K protocol from scratch. The protocol's hybrid architecture combining centralized servers with peer-to-peer transfers creates a robust and scalable file sharing system.

Key implementation priorities:
1. Robust packet parsing with proper error handling
2. Efficient chunk-based downloading with verification
3. Proper connection management for both server and peer connections
4. Integration of both server-based and DHT-based peer discovery

The provided Golang examples demonstrate practical implementation patterns that can be adapted to any programming language while maintaining protocol compatibility with existing ED2K clients.

---

## References

1. eDonkey2000 Protocol Documentation
2. eMule Source Code Analysis
3. Kademlia DHT Specification (Maymounkov & Mazières, 2002)
4. MD4 Hash Algorithm (RFC 1320)

**Appendix A: Protocol Constants**
**Appendix B: Packet Opcode Reference** 
**Appendix C: Tag Type Definitions**
**Appendix D: Error Code Specifications**
