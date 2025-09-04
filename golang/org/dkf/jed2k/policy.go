package jed2k

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

const (
	MaxPeerListSize      = 100
	MinReconnectTimeout  = 10 // seconds
)

// PeerInfo source flags (matching Java constants)
const (
	SourceServer   = 1 << 0
	SourceDHT      = 1 << 1
	SourceIncoming = 1 << 2
	SourceResume   = 1 << 3
)

// Policy manages peers for a transfer - chooses peers for connections
type Policy struct {
	peers       []*Peer
	transfer    *Transfer
	roundRobin  int
	mutex       sync.RWMutex
	rand        *rand.Rand
}

// NewPolicy creates a new policy for a transfer
func NewPolicy(t *Transfer) *Policy {
	return &Policy{
		peers:    make([]*Peer, 0, MaxPeerListSize),
		transfer: t,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// IsConnectCandidate returns true if peer is a candidate for connection
func (p *Policy) IsConnectCandidate(pe *Peer) bool {
	if pe == nil {
		return false
	}
	// Don't connect if peer has active connection, is not connectable, or has too many failures
	return !(pe.HasConnection() || !pe.IsConnectable() || pe.GetFailCount() > 10)
}

// IsEraseCandidate returns true if peer should be erased from the list
func (p *Policy) IsEraseCandidate(pe *Peer) bool {
	if pe.HasConnection() || p.IsConnectCandidate(pe) {
		return false
	}
	return pe.GetFailCount() > 0
}

// GetPeer finds a peer by endpoint
func (p *Policy) GetPeer(endpoint *protocol.Endpoint) *Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	for _, peer := range p.peers {
		if peer.GetEndpoint().Equals(endpoint) {
			return peer
		}
	}
	return nil
}

// AddPeer adds a peer to the policy
func (p *Policy) AddPeer(peer *Peer) (bool, error) {
	if peer == nil {
		return false, fmt.Errorf("peer cannot be nil")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if we need to erase peers first
	if MaxPeerListSize != 0 && len(p.peers) >= MaxPeerListSize {
		p.erasePeersUnlocked()
		if len(p.peers) >= MaxPeerListSize {
			return false, exception.NewJED2KException(exception.PeerLimitExceeded)
		}
	}

	// Find insertion position using binary search
	insertPos := p.findInsertPosition(peer)
	
	// Update peer source flag if peer already exists
	if insertPos < len(p.peers) && p.peers[insertPos].Equals(peer) {
		existingPeer := p.peers[insertPos]
		existingPeer.SetSourceFlag(existingPeer.GetSourceFlag() | peer.GetSourceFlag())
		return false, nil
	}

	// Insert peer at the correct position
	p.peers = append(p.peers, nil)
	copy(p.peers[insertPos+1:], p.peers[insertPos:])
	p.peers[insertPos] = peer

	return true, nil
}

// findInsertPosition finds where to insert a peer to maintain sorted order
func (p *Policy) findInsertPosition(peer *Peer) int {
	return sort.Search(len(p.peers), func(i int) bool {
		return p.peers[i].Compare(peer) >= 0
	})
}

// erasePeersUnlocked removes peers that should be erased (caller must hold lock)
func (p *Policy) erasePeersUnlocked() {
	if MaxPeerListSize == 0 || len(p.peers) == 0 {
		return
	}

	eraseCandidate := -1
	roundRobin := p.rand.Intn(len(p.peers))
	lowWatermark := MaxPeerListSize * 95 / 100
	if lowWatermark == MaxPeerListSize {
		lowWatermark--
	}

	maxIterations := len(p.peers)
	if maxIterations > 300 {
		maxIterations = 300
	}

	for iterations := maxIterations; iterations > 0 && len(p.peers) >= lowWatermark; iterations-- {
		if roundRobin >= len(p.peers) {
			roundRobin = 0
		}

		peer := p.peers[roundRobin]
		current := roundRobin

		if p.IsEraseCandidate(peer) {
			if eraseCandidate == -1 || !p.comparePeerErase(p.peers[eraseCandidate], peer) {
				if p.shouldEraseImmediately(peer) {
					// Remove immediately
					if eraseCandidate > current {
						eraseCandidate--
					}
					p.peers = append(p.peers[:current], p.peers[current+1:]...)
					continue
				} else {
					eraseCandidate = current
				}
			}
		}

		roundRobin++
	}

	// Remove the erase candidate if we found one
	if eraseCandidate != -1 && eraseCandidate < len(p.peers) {
		p.peers = append(p.peers[:eraseCandidate], p.peers[eraseCandidate+1:]...)
	}
}

// comparePeerErase compares two peers for erasing priority (true if first should be erased over second)
func (p *Policy) comparePeerErase(peer1, peer2 *Peer) bool {
	// Prefer to erase peers with more failures
	if peer1.GetFailCount() != peer2.GetFailCount() {
		return peer1.GetFailCount() > peer2.GetFailCount()
	}
	
	// Then prefer to erase peers connected longer ago
	return peer1.GetLastConnected() < peer2.GetLastConnected()
}

// shouldEraseImmediately returns true if peer should be erased immediately
func (p *Policy) shouldEraseImmediately(peer *Peer) bool {
	return (peer.GetSourceFlag() & SourceResume) == SourceResume
}

// FindConnectCandidate finds the best peer to connect to
func (p *Policy) FindConnectCandidate(sessionTime int64) *Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if len(p.peers) == 0 {
		return nil
	}

	candidate := -1
	eraseCandidate := -1
	
	// Try up to twice the number of peers to find a good candidate
	maxIterations := len(p.peers) * 2
	if maxIterations > 300 {
		maxIterations = 300
	}

	for iterations := 0; iterations < maxIterations; iterations++ {
		if p.roundRobin >= len(p.peers) {
			p.roundRobin = 0
		}

		peer := p.peers[p.roundRobin]
		current := p.roundRobin

		// Check if this peer is an erase candidate
		if p.IsEraseCandidate(peer) && (eraseCandidate == -1 || !p.comparePeerErase(p.peers[eraseCandidate], peer)) {
			eraseCandidate = current
		}

		// Check if this peer is a connect candidate
		if p.IsConnectCandidate(peer) {
			// Skip if we already have a better candidate
			if candidate != -1 && p.comparePeersForConnection(p.peers[candidate], peer) {
				p.roundRobin++
				continue
			}
			
			// Skip if peer has a next connection time in the future
			if peer.GetNextConnection() != 0 && peer.GetNextConnection() > sessionTime {
				p.roundRobin++
				continue
			}
			
			// Skip if we need to wait for reconnect timeout (exponential backoff)
			if peer.GetLastConnected() != 0 {
				timeoutSeconds := int64(peer.GetFailCount()+1) * MinReconnectTimeout
				if sessionTime < peer.GetLastConnected()+timeoutSeconds {
					p.roundRobin++
					continue
				}
			}
			
			candidate = current
		}

		p.roundRobin++
	}

	// Remove erase candidate if found
	if eraseCandidate != -1 {
		if candidate > eraseCandidate {
			candidate--
		}
		// Note: In a real implementation, we should remove the peer here
		// For now, we'll just mark it for later cleanup
	}

	if candidate == -1 {
		return nil
	}
	return p.peers[candidate]
}

// comparePeersForConnection compares two peers for connection priority (true if first is better)
func (p *Policy) comparePeersForConnection(peer1, peer2 *Peer) bool {
	// Prefer peers with fewer failures
	if peer1.GetFailCount() != peer2.GetFailCount() {
		return peer1.GetFailCount() < peer2.GetFailCount()
	}
	
	// Then prefer peers connected less recently
	return peer1.GetLastConnected() < peer2.GetLastConnected()
}

// ConnectOnePeer attempts to connect to one peer
func (p *Policy) ConnectOnePeer(sessionTime int64) (bool, error) {
	peer := p.FindConnectCandidate(sessionTime)
	if peer != nil && p.transfer != nil {
		// Get session from transfer (we'd need to modify Transfer to store session reference)
		// For now, we'll simulate the connection attempt
		peer.SetLastConnected(sessionTime)
		
		// Simulate connection success/failure
		if rand.Float64() < 0.7 { // 70% success rate
			// Connection successful - this would be handled by the actual connection logic
			return true, nil
		} else {
			// Connection failed
			peer.SetFailCount(peer.GetFailCount() + 1)
			return false, nil
		}
	}
	return false, nil
}

// ConnectionClosed is called when a peer connection is closed
func (p *Policy) ConnectionClosed(conn *PeerConnection, sessionTime int64) {
	if conn == nil {
		return
	}
	
	peer := conn.GetPeer()
	if peer == nil {
		return
	}
	
	peer.SetConnection(nil)
	peer.SetLastConnected(sessionTime)
	
	if conn.IsFailed() {
		peer.SetFailCount(peer.GetFailCount() + 1)
	}
	
	// Remove peer if it's no longer connectable
	if !peer.IsConnectable() {
		p.removePeer(peer)
	}
}

// removePeer removes a peer from the list
func (p *Policy) removePeer(peerToRemove *Peer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for i, peer := range p.peers {
		if peer.Equals(peerToRemove) {
			p.peers = append(p.peers[:i], p.peers[i+1:]...)
			break
		}
	}
}

// SetConnection sets the connection for a peer
func (p *Policy) SetConnection(peer *Peer, conn *PeerConnection) {
	if conn == nil || peer == nil {
		return
	}
	peer.SetConnection(conn)
}

// NumConnectCandidates returns the number of peers that can be connected to
func (p *Policy) NumConnectCandidates() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	if p.transfer != nil && p.transfer.IsFinished() {
		return 0
	}
	
	count := 0
	for _, peer := range p.peers {
		if p.IsConnectCandidate(peer) && !p.IsEraseCandidate(peer) {
			count++
		}
	}
	return count
}

// NewConnection handles a new incoming connection
func (p *Policy) NewConnection(conn *PeerConnection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	peer := p.GetPeer(conn.GetEndpoint())
	
	if peer != nil {
		// Check for duplicate connection
		if peer.HasConnection() {
			return exception.NewJED2KException(exception.DuplicatePeerConnection)
		}
	} else {
		// Create new peer
		peer = NewPeerWithFlags(conn.GetEndpoint(), false, 0)
		added, err := p.AddPeer(peer)
		if err != nil {
			return err
		}
		if !added {
			return exception.NewJED2KException(exception.DuplicatePeer)
		}
	}
	
	peer.SetConnection(conn)
	conn.SetPeer(peer)
	
	return nil
}

// Size returns the number of peers
func (p *Policy) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.peers)
}

// GetPeers returns a copy of the peers list
func (p *Policy) GetPeers() []*Peer {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	peers := make([]*Peer, len(p.peers))
	copy(peers, p.peers)
	return peers
}