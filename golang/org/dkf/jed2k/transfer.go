package jed2k

import (
	"sync"
	"time"

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
	
	// Peers
	peers             map[string]*PeerInfo
	
	// Control
	paused            bool
	
	// Synchronization
	mutex             sync.RWMutex
	
	// Resume data
	resumeData        ResumeData
}

// NewTransfer creates a new transfer
func NewTransfer(h *hash.Hash, size int64, name, downloadDir string) *Transfer {
	return &Transfer{
		hash:              h,
		size:              size,
		name:              name,
		downloadDirectory: downloadDir,
		createTime:        time.Now(),
		state:             TransferStateQueued,
		peers:             make(map[string]*PeerInfo),
		resumeData:        NewMemoryResumeData(),
	}
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

// GetStatus returns the current transfer status
func (t *Transfer) GetStatus() TransferStatus {
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
	
	return TransferStatus{
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

// updateStats updates transfer statistics (internal method)
func (t *Transfer) updateStats(downloaded, uploaded int64, downloadRate, uploadRate float64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.downloaded = downloaded
	t.uploaded = uploaded
	t.downloadRate = downloadRate
	t.uploadRate = uploadRate
	
	if t.downloaded >= t.size && t.size > 0 {
		t.state = TransferStateCompleted
	} else if !t.paused && t.state != TransferStateError {
		t.state = TransferStateDownloading
	}
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