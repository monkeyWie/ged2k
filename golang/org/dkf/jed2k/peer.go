package jed2k

import (
	"fmt"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// Peer represents information about a peer
type Peer struct {
	lastConnected  int64
	nextConnection int64
	failCount      int
	Connectable    bool
	sourceFlag     int
	connection     *PeerConnection
	endpoint       *protocol.Endpoint
}

// NewPeer creates a new peer with endpoint
func NewPeer(ep *protocol.Endpoint) *Peer {
	if ep == nil {
		panic("endpoint cannot be nil")
	}
	return &Peer{
		endpoint: ep,
	}
}

// NewPeerWithFlags creates a new peer with endpoint, connectable flag and source flag
func NewPeerWithFlags(ep *protocol.Endpoint, conn bool, sourceFlag int) *Peer {
	if ep == nil {
		panic("endpoint cannot be nil")
	}
	return &Peer{
		endpoint:    ep,
		Connectable: conn,
		sourceFlag:  sourceFlag,
	}
}

// HasConnection returns true if peer has an active connection
func (p *Peer) HasConnection() bool {
	return p.connection != nil
}

// GetLastConnected returns the last connected time
func (p *Peer) GetLastConnected() int64 {
	return p.lastConnected
}

// GetNextConnection returns the next connection time
func (p *Peer) GetNextConnection() int64 {
	return p.nextConnection
}

// GetFailCount returns the failure count
func (p *Peer) GetFailCount() int {
	return p.failCount
}

// IsConnectable returns true if peer is connectable
func (p *Peer) IsConnectable() bool {
	return p.Connectable
}

// GetSourceFlag returns the source flag
func (p *Peer) GetSourceFlag() int {
	return p.sourceFlag
}

// GetConnection returns the peer connection
func (p *Peer) GetConnection() *PeerConnection {
	return p.connection
}

// GetEndpoint returns the endpoint
func (p *Peer) GetEndpoint() *protocol.Endpoint {
	return p.endpoint
}

// SetLastConnected sets the last connected time
func (p *Peer) SetLastConnected(lastConnected int64) {
	p.lastConnected = lastConnected
}

// SetNextConnection sets the next connection time
func (p *Peer) SetNextConnection(nextConnection int64) {
	p.nextConnection = nextConnection
}

// SetFailCount sets the failure count
func (p *Peer) SetFailCount(failCount int) {
	p.failCount = failCount
}

// SetConnectable sets the connectable flag
func (p *Peer) SetConnectable(connectable bool) {
	p.Connectable = connectable
}

// SetSourceFlag sets the source flag
func (p *Peer) SetSourceFlag(sourceFlag int) {
	p.sourceFlag = sourceFlag
}

// SetConnection sets the peer connection
func (p *Peer) SetConnection(connection *PeerConnection) {
	p.connection = connection
}

// String returns string representation of peer
func (p *Peer) String() string {
	return fmt.Sprintf("Peer(lastConnected=%d, nextConnection=%d, failCount=%d, connectable=%t, sourceFlag=%d, endpoint=%s)",
		p.lastConnected, p.nextConnection, p.failCount, p.Connectable, p.sourceFlag, p.endpoint)
}

// Equals checks if two peers are equal
func (p *Peer) Equals(other *Peer) bool {
	if other == nil {
		return false
	}
	return p.endpoint.Equals(other.endpoint)
}

// Compare compares two peers by endpoint
func (p *Peer) Compare(other *Peer) int {
	return p.endpoint.CompareTo(other.endpoint)
}