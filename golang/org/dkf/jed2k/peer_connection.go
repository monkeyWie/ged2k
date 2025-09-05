package jed2k

import (
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// PeerConnection represents a connection to a peer
type PeerConnection struct {
	endpoint     *protocol.Endpoint
	peer         *Peer
	session      *Session
	
	// Connection state
	connected    bool
	disconnecting bool
	failed       bool
	
	// Timing
	createTime   int64
	lastReceive  int64
	lastSend     int64
	
	// Statistics
	statistics   *Statistics
	
	// Synchronization
	mutex        sync.RWMutex
}

// NewPeerConnection creates a new peer connection
func NewPeerConnection(endpoint *protocol.Endpoint, session *Session) *PeerConnection {
	now := time.Now().Unix()
	return &PeerConnection{
		endpoint:    endpoint,
		session:     session,
		connected:   false,
		createTime:  now,
		lastReceive: now,
		lastSend:    now,
		statistics:  NewStatistics(),
	}
}

// GetEndpoint returns the connection endpoint
func (pc *PeerConnection) GetEndpoint() *protocol.Endpoint {
	return pc.endpoint
}

// GetPeer returns the associated peer
func (pc *PeerConnection) GetPeer() *Peer {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.peer
}

// SetPeer sets the associated peer
func (pc *PeerConnection) SetPeer(peer *Peer) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.peer = peer
}

// IsConnected returns true if the connection is active
func (pc *PeerConnection) IsConnected() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.connected
}

// IsDisconnecting returns true if the connection is in the process of disconnecting
func (pc *PeerConnection) IsDisconnecting() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.disconnecting
}

// IsFailed returns true if the connection failed
func (pc *PeerConnection) IsFailed() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.failed
}

// GetStatistics returns connection statistics
func (pc *PeerConnection) GetStatistics() *Statistics {
	return pc.statistics
}

// Connect establishes the connection
func (pc *PeerConnection) Connect() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.connected || pc.disconnecting {
		return nil
	}
	
	// Implement actual connection establishment
	// Simulate TCP connection to the peer endpoint
	pc.connected = true
	pc.lastReceive = time.Now().Unix()
	
	// In a real implementation, this would:
	// 1. Create TCP socket to the endpoint
	// 2. Send initial Hello packet
	// 3. Set up packet combiner for ED2K protocol
	// 4. Handle handshake negotiation
	
	return nil
}

// Close closes the peer connection
func (pc *PeerConnection) Close(errorCode exception.ErrorCode) error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.disconnecting {
		return nil
	}
	
	pc.disconnecting = true
	pc.connected = false
	
	if errorCode != exception.NoError {
		pc.failed = true
	}
	
	// Implement actual connection closing
	pc.connected = false
	pc.disconnecting = false
	
	// In a real implementation, this would:
	// 1. Close the TCP socket
	// 2. Abort any pending requests
	// 3. Clean up buffers and packet combiner
	// 4. Update transfer statistics
	
	return nil
}

// SecondTick is called every second to update the connection
func (pc *PeerConnection) SecondTick(tickIntervalMS int64) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.disconnecting {
		return
	}
	
	// Update statistics
	if pc.statistics != nil {
		pc.statistics.SecondTick(tickIntervalMS)
	}
	
	// Check for connection timeout
	if pc.session != nil && pc.session.settings != nil {
		currentTime := time.Now().Unix()
		timeoutSeconds := int64(pc.session.settings.PeerConnectionTimeout)
		
		if currentTime - pc.lastReceive > timeoutSeconds {
			pc.failed = true
			pc.disconnecting = true
			pc.connected = false
		}
	}
}

// UpdateLastReceive updates the last receive time
func (pc *PeerConnection) UpdateLastReceive() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.lastReceive = time.Now().Unix()
}

// UpdateLastSend updates the last send time
func (pc *PeerConnection) UpdateLastSend() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.lastSend = time.Now().Unix()
}