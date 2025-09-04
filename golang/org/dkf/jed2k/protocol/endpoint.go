package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

// Endpoint represents an IP address and port combination
type Endpoint struct {
	ip   uint32
	port uint16
}

// NewEndpoint creates a new endpoint
func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

// NewEndpointFromIPPort creates endpoint from IP and port
func NewEndpointFromIPPort(ip uint32, port uint16) *Endpoint {
	return &Endpoint{ip: ip, port: port}
}

// NewEndpointFromAddr creates endpoint from net.Addr
func NewEndpointFromAddr(addr net.Addr) (*Endpoint, error) {
	switch a := addr.(type) {
	case *net.TCPAddr:
		return NewEndpointFromTCPAddr(a), nil
	case *net.UDPAddr:
		return NewEndpointFromUDPAddr(a), nil
	default:
		return nil, fmt.Errorf("unsupported address type: %T", addr)
	}
}

// NewEndpointFromTCPAddr creates endpoint from TCPAddr
func NewEndpointFromTCPAddr(addr *net.TCPAddr) *Endpoint {
	ip := ipToUint32(addr.IP)
	return &Endpoint{ip: ip, port: uint16(addr.Port)}
}

// NewEndpointFromUDPAddr creates endpoint from UDPAddr
func NewEndpointFromUDPAddr(addr *net.UDPAddr) *Endpoint {
	ip := ipToUint32(addr.IP)
	return &Endpoint{ip: ip, port: uint16(addr.Port)}
}

// NewEndpointFromString creates endpoint from "ip:port" string
func NewEndpointFromString(addr string) (*Endpoint, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", host)
	}
	
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}
	
	return &Endpoint{
		ip:   ipToUint32(ip),
		port: uint16(port),
	}, nil
}

// ipToUint32 converts IP to uint32 in network byte order
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip)
}

// uint32ToIP converts uint32 to IP in network byte order
func uint32ToIP(ip uint32) net.IP {
	result := make(net.IP, 4)
	binary.BigEndian.PutUint32(result, ip)
	return result
}

// Get deserializes from buffer
func (e *Endpoint) Get(src *bytes.Buffer) error {
	if src.Len() < 6 {
		return fmt.Errorf("buffer underflow: need 6 bytes, have %d", src.Len())
	}
	
	// Read IP in little endian 
	err := binary.Read(src, binary.LittleEndian, &e.ip)
	if err != nil {
		return err
	}
	
	// Read port in little endian
	err = binary.Read(src, binary.LittleEndian, &e.port)
	return err
}

// Put serializes to buffer
func (e *Endpoint) Put(dst *bytes.Buffer) error {
	// Write IP in little endian
	err := binary.Write(dst, binary.LittleEndian, e.ip)
	if err != nil {
		return err
	}
	
	// Write port in little endian
	return binary.Write(dst, binary.LittleEndian, e.port)
}

// BytesCount returns the number of bytes this type uses
func (e *Endpoint) BytesCount() int {
	return 6 // 4 bytes for IP + 2 bytes for port
}

// Assign sets values from another endpoint
func (e *Endpoint) Assign(other *Endpoint) *Endpoint {
	e.ip = other.ip
	e.port = other.port
	return e
}

// AssignIPPort sets IP and port values
func (e *Endpoint) AssignIPPort(ip uint32, port uint16) *Endpoint {
	e.ip = ip
	e.port = port
	return e
}

// Defined checks if endpoint has valid values
func (e *Endpoint) Defined() bool {
	return e.ip != 0 && e.port != 0
}

// IP returns the IP address as uint32
func (e *Endpoint) IP() uint32 {
	return e.ip
}

// Port returns the port
func (e *Endpoint) Port() uint16 {
	return e.port
}

// IPNet returns the IP as net.IP
func (e *Endpoint) IPNet() net.IP {
	return uint32ToIP(e.ip)
}

// TCPAddr returns a net.TCPAddr
func (e *Endpoint) TCPAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   e.IPNet(),
		Port: int(e.port),
	}
}

// UDPAddr returns a net.UDPAddr
func (e *Endpoint) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   e.IPNet(),
		Port: int(e.port),
	}
}

// String returns string representation
func (e *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.IPNet().String(), e.port)
}

// Equals checks equality
func (e *Endpoint) Equals(other *Endpoint) bool {
	return e.ip == other.ip && e.port == other.port
}

// CompareTo compares with another endpoint
func (e *Endpoint) CompareTo(other *Endpoint) int {
	if e.ip < other.ip {
		return -1
	} else if e.ip > other.ip {
		return 1
	}
	
	if e.port < other.port {
		return -1
	} else if e.port > other.port {
		return 1
	}
	
	return 0
}