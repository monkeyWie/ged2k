package jed2k

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// Utils provides utility functions for ed2k protocol
type Utils struct{}

// Byte2String converts byte array to uppercase hex string
func Byte2String(value []byte) string {
	if value == nil {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(value))
}

// Byte2StringSingle converts single byte to uppercase hex string
func Byte2StringSingle(value byte) string {
	return Byte2String([]byte{value})
}

// Sizeof functions for basic types
func SizeofByte(value byte) int     { return 1 }
func SizeofInt16(value int16) int   { return 2 }
func SizeofInt32(value int32) int   { return 4 }
func SizeofFloat32(value float32) int { return 4 }
func SizeofInt64(value int64) int   { return 8 }
func SizeofBool(value bool) int     { return 1 }

// SizeofSerializable returns bytes count for serializable object
func SizeofSerializable(s protocol.Serializable) int {
	return s.BytesCount()
}

// Int2Address converts little-endian int to net.IP
// ip is interpreted as network order (big-endian) with MSB at position 0
func Int2Address(ip uint32) net.IP {
	return net.IPv4(
		byte(ip&0xff),
		byte((ip>>8)&0xff),
		byte((ip>>16)&0xff),
		byte((ip>>24)&0xff),
	)
}

// IP2String converts IP integer to dotted decimal string
func IP2String(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip&0xff,
		(ip>>8)&0xff,
		(ip>>16)&0xff,
		(ip>>24)&0xff)
}

// String2IP converts dotted decimal string to IP integer
func String2IP(s string) (uint32, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("invalid IP format: %s", s)
	}
	
	var raw [4]int
	for i, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("invalid IP part '%s': %v", part, err)
		}
		if val < 0 || val > 255 {
			return 0, fmt.Errorf("IP part out of range: %d", val)
		}
		raw[i] = val
	}
	
	return uint32(raw[0]) |
		uint32(raw[1]<<8) |
		uint32(raw[2]<<16) |
		uint32(raw[3]<<24), nil
}

// Htonl converts host byte order to network byte order (big-endian)
func HtonlBytes(order []byte) uint32 {
	if len(order) != 4 {
		panic("order must be 4 bytes")
	}
	return uint32(order[0])<<24 |
		uint32(order[1])<<16 |
		uint32(order[2])<<8 |
		uint32(order[3])
}

// HtonlUint32 converts uint32 from host to network byte order
func HtonlUint32(ip uint32) uint32 {
	return ((ip & 0x000000FF) << 24) |
		((ip & 0x0000FF00) << 8) |
		((ip >> 8) & 0x0000FF00) |
		((ip >> 24) & 0xFF)
}

// PackToNetworkByteOrder packs bytes to network byte order
// order is bytes in network byte order - MSB at position 0
func PackToNetworkByteOrder(order []byte) uint32 {
	if len(order) != 4 {
		panic("order must be 4 bytes")
	}
	return uint32(order[3])<<24 |
		uint32(order[2])<<16 |
		uint32(order[1])<<8 |
		uint32(order[0])
}

// Ntohl converts network byte order to host byte order
func Ntohl(ip uint32) uint32 {
	raw := []byte{
		byte(ip & 0xff),
		byte((ip >> 8) & 0xff),
		byte((ip >> 16) & 0xff),
		byte((ip >> 24) & 0xff),
	}
	return HtonlBytes(raw)
}

// IsLocalAddress checks if endpoint represents a local/private address
func IsLocalAddress(ep *protocol.Endpoint) bool {
	host := Ntohl(ep.IP())
	return (host&0xff000000) == 0x0a000000 || // 10.x.x.x
		(host&0xfff00000) == 0xac100000 || // 172.16.x.x
		(host&0xffff0000) == 0xc0a80000 || // 192.168.x.x
		(host&0xffff0000) == 0xa9fe0000 || // 169.254.x.x
		(host&0xff000000) == 0x7f000000 // 127.x.x.x
}

// LowPart returns lower 32 bits of 64-bit value
func LowPart(value uint64) uint32 {
	return uint32(value)
}

// HiPart returns upper 32 bits of 64-bit value  
func HiPart(value uint64) uint32 {
	return uint32(value >> 32)
}

// MakeFullED2KVersion creates full ed2k version from components
func MakeFullED2KVersion(clientID, a, b, c uint32) uint64 {
	return (uint64(clientID) << 24) |
		(uint64(a) << 17) |
		(uint64(b) << 10) |
		(uint64(c) << 7)
}

// UAgent2CSoft converts user agent hash to client software type
func UAgent2CSoft(hash *protocol.Hash) ClientSoftware {
	bytes := hash.Bytes()
	
	if len(bytes) < 15 {
		return SOUnknown
	}
	
	if bytes[5] == 13 && bytes[14] == 110 {
		return SOOldEmule
	}
	
	if bytes[5] == 14 && bytes[14] == 111 {
		return SOEmule
	}
	
	if bytes[5] == 'M' && bytes[14] == 'L' {
		return SOMLDonkey
	}
	
	if bytes[5] == 'L' && bytes[14] == 'K' {
		return SOLibED2K
	}
	
	if bytes[5] == 'Q' && bytes[14] == 'M' {
		return SOQMule
	}
	
	return SOUnknown
}