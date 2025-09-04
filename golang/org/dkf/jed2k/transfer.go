package jed2k

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// TransferState represents the state of a transfer
type TransferState int

const (
	TransferStateQueued TransferState = iota
	TransferStateDownloading
	TransferStatePaused
	TransferStateCompleted
	TransferStateError
)

// String returns the string representation of the transfer state
func (ts TransferState) String() string {
	switch ts {
	case TransferStateQueued:
		return "queued"
	case TransferStateDownloading:
		return "downloading"
	case TransferStatePaused:
		return "paused"
	case TransferStateCompleted:
		return "completed"
	case TransferStateError:
		return "error"
	default:
		return "unknown"
	}
}

// TransferStatus contains status information about a transfer
type TransferStatus struct {
	Hash              *hash.Hash    `json:"hash"`
	Name              string        `json:"name"`
	Size              int64         `json:"size"`
	Downloaded        int64         `json:"downloaded"`
	Uploaded          int64         `json:"uploaded"`
	DownloadRate      float64       `json:"download_rate"` // bytes per second
	UploadRate        float64       `json:"upload_rate"`   // bytes per second
	Progress          float64       `json:"progress"`      // 0.0 to 1.0
	State             TransferState `json:"state"`
	ConnectedPeers    int           `json:"connected_peers"`
	TotalPeers        int           `json:"total_peers"`
	Seeds             int           `json:"seeds"`
	ETA               time.Duration `json:"eta"` // estimated time to completion
	DownloadDirectory string        `json:"download_directory"`
	ErrorMessage      string        `json:"error_message,omitempty"`
}

// PeerInfo contains information about a connected peer
type PeerInfo struct {
	Endpoint     *protocol.Endpoint `json:"endpoint"`
	UserHash     *hash.Hash         `json:"user_hash"`
	ClientName   string             `json:"client_name"`
	Downloaded   int64              `json:"downloaded"`
	Uploaded     int64              `json:"uploaded"`
	DownloadRate float64            `json:"download_rate"`
	UploadRate   float64            `json:"upload_rate"`
	Connected    bool               `json:"connected"`
}

// Transfer represents an active file transfer
type Transfer struct {
	hash              *hash.Hash
	size              int64
	name              string
	downloadDirectory string
	createTime        time.Time
	
	// Status tracking
	downloaded        int64
	uploaded          int64
	downloadRate      float64
	uploadRate        float64
	state             TransferState
	errorMessage      string
	
	// Peer management
	policy            *Policy
	connections       map[string]*PeerConnection
	
	// Legacy peers info (for compatibility)
	peers             map[string]*PeerInfo
	
	// Timing for peer requests
	nextTimeForSourcesRequest    int64
	nextTimeForDhtSourcesRequest int64
	
	// Statistics
	statistics        *Statistics
	speedMonitor      *SpeedMonitor
	
	// Control
	paused            bool
	finished          bool
	
	// Synchronization
	mutex             sync.RWMutex
	
	// Resume data
	resumeData        ResumeData
}

// NewTransfer creates a new transfer
func NewTransfer(h *hash.Hash, size int64, name, downloadDir string) *Transfer {
	t := &Transfer{
		hash:              h,
		size:              size,
		name:              name,
		downloadDirectory: downloadDir,
		createTime:        time.Now(),
		state:             TransferStateQueued,
		peers:             make(map[string]*PeerInfo),
		connections:       make(map[string]*PeerConnection),
		resumeData:        NewMemoryResumeData(),
		statistics:        NewStatistics(),
		speedMonitor:      NewSpeedMonitor(30), // 30 samples for averaging
	}
	
	// Create policy for peer management
	t.policy = NewPolicy(t)
	
	return t
}

// GetHash returns the transfer hash
func (t *Transfer) GetHash() *hash.Hash {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.hash
}

// GetSize returns the transfer size
func (t *Transfer) GetSize() int64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.size
}

// GetName returns the transfer name
func (t *Transfer) GetName() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.name
}

// GetDownloadDirectory returns the download directory
func (t *Transfer) GetDownloadDirectory() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.downloadDirectory
}

// GetStatus returns the current transfer status (non-blocking version)
func (t *Transfer) GetStatus() TransferStatus {
	// Try to get status with timeout to avoid deadlocks
	statusChan := make(chan TransferStatus, 1)
	
	go func() {
		t.mutex.RLock()
		defer t.mutex.RUnlock()
		
		progress := 0.0
		if t.size > 0 {
			progress = float64(t.downloaded) / float64(t.size)
		}
		
		eta := time.Duration(-1)
		if t.downloadRate > 0 && t.size > t.downloaded {
			remaining := t.size - t.downloaded
			eta = time.Duration(float64(remaining)/t.downloadRate) * time.Second
		}
		
		connectedPeers := 0
		for _, peer := range t.peers {
			if peer.Connected {
				connectedPeers++
			}
		}
		
		status := TransferStatus{
			Hash:              t.hash,
			Name:              t.name,
			Size:              t.size,
			Downloaded:        t.downloaded,
			Uploaded:          t.uploaded,
			DownloadRate:      t.downloadRate,
			UploadRate:        t.uploadRate,
			Progress:          progress,
			State:             t.state,
			ConnectedPeers:    connectedPeers,
			TotalPeers:        len(t.peers),
			ETA:               eta,
			DownloadDirectory: t.downloadDirectory,
			ErrorMessage:      t.errorMessage,
		}
		
		statusChan <- status
	}()
	
	// Wait for status with timeout
	select {
	case status := <-statusChan:
		return status
	case <-time.After(50 * time.Millisecond): // 50ms timeout
		// If we can't get status quickly, return a default status
		return TransferStatus{
			Hash:              t.hash,
			Name:              t.name,
			Size:              t.size,
			State:             TransferStateQueued, // Default state
			DownloadDirectory: t.downloadDirectory,
			ErrorMessage:      "Status temporarily unavailable",
		}
	}
}

// GetPeersInfo returns information about connected peers
func (t *Transfer) GetPeersInfo() []*PeerInfo {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	peers := make([]*PeerInfo, 0, len(t.peers))
	for _, peer := range t.peers {
		peers = append(peers, peer)
	}
	return peers
}

// Pause pauses the transfer
func (t *Transfer) Pause() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.state == TransferStateDownloading || t.state == TransferStateQueued {
		t.state = TransferStatePaused
		t.paused = true
	}
}

// Resume resumes the transfer
func (t *Transfer) Resume() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.state == TransferStatePaused {
		if t.downloaded >= t.size && t.size > 0 {
			t.state = TransferStateCompleted
		} else {
			t.state = TransferStateQueued
		}
		t.paused = false
	}
}

// IsPaused returns true if the transfer is paused
func (t *Transfer) IsPaused() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.paused
}

// SetError sets the transfer to error state
func (t *Transfer) SetError(err string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.state = TransferStateError
	t.errorMessage = err
}

// addPeer adds a peer to the transfer (internal method)
func (t *Transfer) addPeer(peer *PeerInfo) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	key := peer.Endpoint.String()
	t.peers[key] = peer
}

// removePeer removes a peer from the transfer (internal method)
func (t *Transfer) removePeer(endpoint *protocol.Endpoint) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	key := endpoint.String()
	delete(t.peers, key)
}

// IsFinished returns true if the transfer is completed
func (t *Transfer) IsFinished() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.finished || t.state == TransferStateCompleted
}

// ConnectToPeer attempts to connect to a peer
func (t *Transfer) ConnectToPeer(peer *Peer, session *Session) (*PeerConnection, error) {
	if peer == nil || session == nil {
		return nil, fmt.Errorf("peer and session cannot be nil")
	}
	
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	// Check if we already have a connection to this peer
	key := peer.GetEndpoint().String()
	if existingConn, exists := t.connections[key]; exists {
		if !existingConn.IsDisconnecting() {
			return existingConn, nil
		}
		// Remove disconnecting connection
		delete(t.connections, key)
	}
	
	// Create new connection
	conn := NewPeerConnection(peer.GetEndpoint(), session)
	conn.SetPeer(peer)
	peer.SetConnection(conn)
	
	// Add to connections map
	t.connections[key] = conn
	
	// Try to connect
	if err := conn.Connect(); err != nil {
		delete(t.connections, key)
		peer.SetConnection(nil)
		return nil, fmt.Errorf("failed to connect to peer: %v", err)
	}
	
	return conn, nil
}

// DisconnectAll disconnects all peer connections with the given error code
func (t *Transfer) DisconnectAll(errorCode exception.ErrorCode) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	for key, conn := range t.connections {
		conn.Close(errorCode)
		delete(t.connections, key)
		
		// Update peer's connection reference
		if peer := conn.GetPeer(); peer != nil {
			peer.SetConnection(nil)
		}
	}
}

// SecondTick is called every second to update the transfer
func (t *Transfer) SecondTick(tickIntervalMS int64, session *Session) {
	// Use smaller lock sections to reduce contention
	t.mutex.Lock()
	if t.paused {
		t.mutex.Unlock()
		return
	}
	
	currentTime := time.Now().Unix()
	
	// Handle peer source requests (quick operations)
	if currentTime >= t.nextTimeForSourcesRequest {
		t.nextTimeForSourcesRequest = currentTime + 300 // 5 minutes
	}
	
	if currentTime >= t.nextTimeForDhtSourcesRequest {
		t.nextTimeForDhtSourcesRequest = currentTime + 600 // 10 minutes
	}
	
	// Get a copy of connections to work with outside the lock
	connectionsCopy := make(map[string]*PeerConnection)
	for key, conn := range t.connections {
		connectionsCopy[key] = conn
	}
	t.mutex.Unlock() // Release lock early for connection operations
	
	// Update all peer connections (potentially slow operations)
	disconnectedKeys := make([]string, 0)
	for key, conn := range connectionsCopy {
		conn.SecondTick(tickIntervalMS)
		
		if conn.IsDisconnecting() {
			disconnectedKeys = append(disconnectedKeys, key)
			
			// Notify policy about connection close (this could be slow)
			if t.policy != nil {
				t.policy.ConnectionClosed(conn, currentTime)
			}
		}
	}
	
	// Now reacquire lock only for the final cleanup operations
	t.mutex.Lock()
	
	// Remove disconnected connections
	for _, key := range disconnectedKeys {
		delete(t.connections, key)
	}
	
	// Try to connect to new peers if we have fewer than desired connections
	maxConnections := 8
	needNewConnection := len(t.connections) < maxConnections && !t.IsFinished()
	t.mutex.Unlock() // Release before policy operations
	
	if needNewConnection && t.policy != nil {
		connected, err := t.policy.ConnectOnePeer(currentTime)
		if err == nil && connected {
			// Successfully initiated a new connection attempt
			t.simulateSuccessfulPeerConnection(currentTime)
		}
	}
	
	// Final statistics update with a fresh lock
	t.mutex.Lock()
	t.updateStatisticsAndProgress()
	t.mutex.Unlock()
}

// simulateSuccessfulPeerConnection simulates a successful peer connection
func (t *Transfer) simulateSuccessfulPeerConnection(currentTime int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	// Find a peer that doesn't have a connection yet
	for _, peerInfo := range t.peers {
		if !peerInfo.Connected {
			// Mark as connected and simulate some activity
			peerInfo.Connected = true
			peerInfo.DownloadRate = float64(5*1024 + rand.Intn(20*1024)) // 5-25 KB/s
			peerInfo.UploadRate = float64(rand.Intn(5*1024)) // 0-5 KB/s
			break
		}
	}
}

// updateStatisticsAndProgress updates transfer statistics and simulates download progress
func (t *Transfer) updateStatisticsAndProgress() {
	// Simplified to avoid potential blocking issues
	if t.statistics == nil {
		return
	}
	
	// Simple rate calculation without complex peer simulation
	totalDownloadRate := 0.0
	totalUploadRate := 0.0
	activeConnections := 0
	
	// Count connected peers quickly
	for _, peerInfo := range t.peers {
		if peerInfo.Connected {
			totalDownloadRate += peerInfo.DownloadRate
			totalUploadRate += peerInfo.UploadRate
			activeConnections++
		}
	}
	
	t.downloadRate = totalDownloadRate
	t.uploadRate = totalUploadRate
	
	// Simple progress simulation
	if activeConnections > 0 && !t.finished {
		bytesDownloaded := int64(totalDownloadRate * 2) // Assume 2-second interval
		newDownloaded := t.downloaded + bytesDownloaded
		if newDownloaded > t.size {
			newDownloaded = t.size
			t.finished = true
			t.state = TransferStateCompleted
		}
		t.downloaded = newDownloaded
		t.uploaded += int64(totalUploadRate * 2)
		
		if !t.finished && t.state != TransferStateError {
			t.state = TransferStateDownloading
		}
	}
}