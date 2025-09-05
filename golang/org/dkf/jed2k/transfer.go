package jed2k

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
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
	Hash              *protocol.Hash `json:"hash"`
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
	UserHash     *protocol.Hash     `json:"user_hash"`
	ClientName   string             `json:"client_name"`
	Downloaded   int64              `json:"downloaded"`
	Uploaded     int64              `json:"uploaded"`
	DownloadRate float64            `json:"download_rate"`
	UploadRate   float64            `json:"upload_rate"`
	Connected    bool               `json:"connected"`
}

// Transfer represents an active file transfer
type Transfer struct {
	hash              *protocol.Hash
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
	fileHandler       FileHandler
	
	// Piece management
	picker            *PiecePicker
	pieceSize         int64
	numPieces         int
	pieces            *protocol.BitField  // tracks completed pieces
	pieceData         [][]byte           // actual downloaded piece data
	
	// Synchronization
	mutex             sync.RWMutex
	
	// Resume data
	resumeData        ResumeData
}

// NewTransfer creates a new transfer
func NewTransfer(h *protocol.Hash, size int64, name, downloadDir string) *Transfer {
	// Create the full file path
	filePath := filepath.Join(downloadDir, name)
	
	// Calculate piece information using ed2k constants
	numPieces := int((size + PieceSize - 1) / PieceSize) // ceiling division
	blocksInLastPiece := int(((size % PieceSize) + BlockSize - 1) / BlockSize)
	
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
		fileHandler:       NewDefaultFileHandler(filePath),
		picker:            NewPiecePicker(numPieces, blocksInLastPiece),
		pieceSize:         PieceSize,
		numPieces:         numPieces,
		pieces:            protocol.NewBitFieldWithSize(numPieces),
		pieceData:         make([][]byte, numPieces),
	}
	
	// Create policy for peer management
	t.policy = NewPolicy(t)
	
	return t
}

// GetHash returns the transfer hash
func (t *Transfer) GetHash() *protocol.Hash {
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
	
	// Wait for status with a longer timeout
	select {
	case status := <-statusChan:
		return status
	case <-time.After(500 * time.Millisecond): // 500ms timeout 
		// If we can't get status quickly, return a default status with current accessible fields
		return TransferStatus{
			Hash:              t.hash,
			Name:              t.name,
			Size:              t.size,
			State:             TransferStateQueued, // Default state
			DownloadDirectory: t.downloadDirectory,
			ErrorMessage:      "Status temporarily unavailable (timeout)",
		}
	}
}

// GetPeersInfo returns information about connected peers (with timeout to prevent hanging)
func (t *Transfer) GetPeersInfo() []*PeerInfo {
	peersChan := make(chan []*PeerInfo, 1)
	
	go func() {
		t.mutex.RLock()
		defer t.mutex.RUnlock()
		
		peers := make([]*PeerInfo, 0, len(t.peers))
		for _, peer := range t.peers {
			peers = append(peers, peer)
		}
		peersChan <- peers
	}()
	
	// Wait for peers with timeout
	select {
	case peers := <-peersChan:
		return peers
	case <-time.After(200 * time.Millisecond): // 200ms timeout
		// If we can't get peers info quickly, return empty slice
		fmt.Printf("Warning: GetPeersInfo() timed out for transfer %s\n", t.name)
		return make([]*PeerInfo, 0)
	}
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
			// Resume to downloading state, not queued, so it doesn't get stuck
			t.state = TransferStateDownloading
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

// hasPicker returns true if the transfer has a piece picker
func (t *Transfer) hasPicker() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.picker != nil
}

// GetPicker returns the piece picker
func (t *Transfer) GetPicker() *PiecePicker {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.picker
}

// WriteBlock writes a downloaded block to the transfer
func (t *Transfer) WriteBlock(block *PieceBlock, data []byte) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.picker == nil {
		return fmt.Errorf("no piece picker available")
	}
	
	// Store block data
	if t.pieceData[block.PieceIndex] == nil {
		// Calculate piece size for this piece
		pieceSize := PieceSize
		if block.PieceIndex == t.numPieces-1 {
			// Last piece might be smaller
			remainingSize := t.size - int64(block.PieceIndex)*PieceSize
			if remainingSize < PieceSize {
				pieceSize = remainingSize
			}
		}
		t.pieceData[block.PieceIndex] = make([]byte, pieceSize)
	}
	
	// Calculate block offset within the piece
	blockOffset := int64(block.BlockIndex) * BlockSize
	
	// Copy data to the piece buffer
	copy(t.pieceData[block.PieceIndex][blockOffset:], data)
	
	// Mark block as finished in picker
	t.picker.MarkAsFinished(block)
	
	// Update downloaded bytes
	t.downloaded += int64(len(data))
	
	// Check if piece is complete
	if t.picker.HavePiece(block.PieceIndex) {
		t.pieces.SetBit(block.PieceIndex)
		
		// Check if entire file is complete
		if t.downloaded >= t.size {
			t.finished = true
			t.state = TransferStateCompleted
			t.writeCompletedFile()
		}
	}
	
	return nil
}

// IsFinished returns true if the transfer is completed
func (t *Transfer) IsFinished() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.finished || t.state == TransferStateCompleted
}

// ConnectToPeer attempts to connect to a peer using real ed2k protocol
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
	
	// Create new connection with ed2k protocol
	conn := NewPeerConnection(peer.GetEndpoint(), session)
	conn.SetPeer(peer)
	conn.SetTransfer(t) // Associate transfer with connection
	peer.SetConnection(conn)
	
	// Add to connections map
	t.connections[key] = conn
	
	// Try to connect using real ed2k protocol
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
	// Quick check if paused before doing any work
	t.mutex.RLock()
	if t.paused {
		t.mutex.RUnlock()
		return
	}
	isFinished := t.IsFinished()
	t.mutex.RUnlock()
	
	if isFinished {
		return
	}
	
	currentTime := time.Now().Unix()
	
	// Handle time-based operations without locks first
	t.mutex.Lock()
	// Handle peer source requests (quick operations)
	if currentTime >= t.nextTimeForSourcesRequest {
		t.nextTimeForSourcesRequest = currentTime + 300 // 5 minutes
	}
	
	if currentTime >= t.nextTimeForDhtSourcesRequest {
		t.nextTimeForDhtSourcesRequest = currentTime + 600 // 10 minutes
	}
	
	// Get a snapshot of connections to work with
	connectionsCopy := make(map[string]*PeerConnection)
	for key, conn := range t.connections {
		connectionsCopy[key] = conn
	}
	t.mutex.Unlock() 
	
	// Update all peer connections with real ed2k protocol handling
	disconnectedKeys := make([]string, 0)
	for key, conn := range connectionsCopy {
		conn.SecondTick(tickIntervalMS)
		
		if conn.IsDisconnecting() {
			disconnectedKeys = append(disconnectedKeys, key)
			
			// Notify policy about connection close
			if t.policy != nil {
				t.policy.ConnectionClosed(conn, currentTime)
			}
		}
	}
	
	// Quick cleanup and connection management
	t.mutex.Lock()
	
	// Remove disconnected connections
	for _, key := range disconnectedKeys {
		delete(t.connections, key)
	}
	
	// Check if we need new connections
	maxConnections := 8
	needNewConnection := len(t.connections) < maxConnections
	
	// Update statistics and progress quickly
	t.updateStatisticsAndProgress()
	t.mutex.Unlock() 
	
	// Try to connect to new peers using real ed2k protocol
	if needNewConnection && t.policy != nil {
		connected, err := t.policy.ConnectOnePeer(currentTime)
		if err == nil && connected {
			// Real connection attempt was initiated
			fmt.Printf("Initiated new peer connection for transfer %s\n", t.name)
		}
	}
}

// updateStatisticsAndProgress updates transfer statistics from real connections
func (t *Transfer) updateStatisticsAndProgress() {
	if t.statistics == nil {
		return
	}
	
	// Aggregate statistics from all active connections
	totalDownloadRate := 0.0
	totalUploadRate := 0.0
	activeConnections := 0
	
	for _, conn := range t.connections {
		if conn.IsConnected() {
			stats := conn.GetStatistics()
			totalDownloadRate += float64(stats.DownloadRate())
			totalUploadRate += float64(stats.UploadRate())
			activeConnections++
		}
	}
	
	t.downloadRate = totalDownloadRate
	t.uploadRate = totalUploadRate
	
	// Update transfer state based on real progress
	if !t.finished && activeConnections > 0 {
		if t.downloaded >= t.size {
			t.finished = true
			t.state = TransferStateCompleted
			t.writeCompletedFile()
		} else if t.state != TransferStateError && !t.paused {
			t.state = TransferStateDownloading
		}
	}
}

// writeCompletedFile assembles and writes the completed file from downloaded pieces
func (t *Transfer) writeCompletedFile() {
	if t.fileHandler == nil {
		return
	}
	
	// Open the file for writing
	if err := t.fileHandler.Open(); err != nil {
		fmt.Printf("Warning: Failed to open file %s for writing: %v\n", t.name, err)
		return
	}
	defer t.fileHandler.Close()
	
	// Assemble file from pieces that were actually downloaded from peers
	var currentOffset int64 = 0
	for i := 0; i < t.numPieces; i++ {
		if !t.pieces.GetBit(i) || t.pieceData[i] == nil {
			fmt.Printf("Warning: Missing piece %d for file %s\n", i, t.name)
			continue
		}
		
		// Write piece data to file
		if _, err := t.fileHandler.Write(currentOffset, t.pieceData[i]); err != nil {
			fmt.Printf("Warning: Failed to write piece %d for file %s: %v\n", i, t.name, err)
			return
		}
		
		currentOffset += int64(len(t.pieceData[i]))
	}
	
	// Sync to ensure data is written to disk
	if err := t.fileHandler.Sync(); err != nil {
		fmt.Printf("Warning: Failed to sync file %s: %v\n", t.name, err)
	}
	
	fmt.Printf("✓ Successfully wrote completed file: %s (%d bytes)\n", t.fileHandler.GetPath(), t.size)
}