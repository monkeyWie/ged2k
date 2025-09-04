package jed2k

import (
	"fmt"
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// TransferHandle provides a handle to control and monitor a transfer
type TransferHandle struct {
	transfer *Transfer
	session  *Session
}

// NewTransferHandle creates a new transfer handle
func NewTransferHandle(transfer *Transfer, session *Session) *TransferHandle {
	return &TransferHandle{
		transfer: transfer,
		session:  session,
	}
}

// IsValid returns true if the handle is valid
func (th *TransferHandle) IsValid() bool {
	return th.transfer != nil && th.session != nil
}

// GetHash returns the transfer hash
func (th *TransferHandle) GetHash() *hash.Hash {
	if !th.IsValid() {
		return nil
	}
	return th.transfer.GetHash()
}

// GetSize returns the transfer size
func (th *TransferHandle) GetSize() int64 {
	if !th.IsValid() {
		return 0
	}
	return th.transfer.GetSize()
}

// GetName returns the transfer name
func (th *TransferHandle) GetName() string {
	if !th.IsValid() {
		return ""
	}
	return th.transfer.GetName()
}

// GetStatus returns the current transfer status
func (th *TransferHandle) GetStatus() TransferStatus {
	if !th.IsValid() {
		return TransferStatus{}
	}
	return th.transfer.GetStatus()
}

// GetPeersInfo returns information about connected peers
func (th *TransferHandle) GetPeersInfo() []*PeerInfo {
	if !th.IsValid() {
		return nil
	}
	return th.transfer.GetPeersInfo()
}

// Pause pauses the transfer
func (th *TransferHandle) Pause() error {
	if !th.IsValid() {
		return fmt.Errorf("invalid transfer handle")
	}
	
	th.transfer.Pause()
	
	// Save resume data when pausing
	if err := th.session.persistenceManager.SaveTransfer(th.transfer); err != nil {
		return fmt.Errorf("failed to save resume data: %v", err)
	}
	
	return nil
}

// Resume resumes the transfer
func (th *TransferHandle) Resume() error {
	if !th.IsValid() {
		return fmt.Errorf("invalid transfer handle")
	}
	
	th.transfer.Resume()
	return nil
}

// Remove removes the transfer from the session
func (th *TransferHandle) Remove(deleteFiles bool) error {
	if !th.IsValid() {
		return fmt.Errorf("invalid transfer handle")
	}
	
	return th.session.RemoveTransfer(th.transfer.GetHash(), deleteFiles)
}

// Session represents the main ed2k client session
type Session struct {
	settings            *Settings
	transfers           map[string]*Transfer
	transferHandles     map[string]*TransferHandle
	serverList          *ServerList
	nodesData           *NodesData
	persistenceManager  *PersistenceManager
	
	// State
	running             bool
	connected           bool
	
	// Synchronization
	mutex               sync.RWMutex
	
	// Shutdown
	done                chan struct{}
}

// NewSession creates a new ed2k session
func NewSession(settings *Settings) *Session {
	if settings == nil {
		settings = NewDefaultSettings()
	}
	
	// Create persistence manager based on settings
	var resumeData ResumeData
	if settings.ResumeDataDirectory != "" {
		resumeData = NewDiskResumeData(settings.ResumeDataDirectory)
	} else {
		resumeData = NewMemoryResumeData()
	}
	
	session := &Session{
		settings:           settings,
		transfers:          make(map[string]*Transfer),
		transferHandles:    make(map[string]*TransferHandle),
		serverList:         NewServerList(),
		nodesData:          NewNodesData(),
		persistenceManager: NewPersistenceManager(resumeData),
		done:               make(chan struct{}),
	}
	
	return session
}

// Start starts the session
func (s *Session) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.running {
		return fmt.Errorf("session already running")
	}
	
	s.running = true
	
	// Start background routines
	go s.mainLoop()
	
	return nil
}

// Stop stops the session
func (s *Session) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.running {
		return nil
	}
	
	s.running = false
	close(s.done)
	
	// Save all transfer resume data
	for _, transfer := range s.transfers {
		s.persistenceManager.SaveTransfer(transfer)
	}
	
	return nil
}

// IsRunning returns true if the session is running
func (s *Session) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// AddServer adds a server to the server list
func (s *Session) AddServer(endpoint string, name string) error {
	// TODO: Parse endpoint string and add to server list
	return nil
}

// LoadServerList loads servers from a server.met URL
func (s *Session) LoadServerList(url string) error {
	return s.serverList.LoadFromMet(url)
}

// LoadNodesList loads Kademlia nodes from a nodes.dat URL
func (s *Session) LoadNodesList(url string) error {
	return s.nodesData.LoadFromDat(url)
}

// AddTransfer adds a new transfer to the session
func (s *Session) AddTransfer(params *AddTransferParams) (*TransferHandle, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.running {
		return nil, fmt.Errorf("session not running")
	}
	
	hashStr := params.Hash.String()
	
	// Check if transfer already exists
	if _, exists := s.transfers[hashStr]; exists {
		return nil, fmt.Errorf("transfer already exists")
	}
	
	// Create new transfer
	transfer := NewTransfer(params.Hash, params.Size, params.Name, params.DownloadDirectory)
	
	// Apply resume data if provided
	if params.ResumeData != nil {
		// TODO: Apply resume data to transfer
	}
	
	// Set initial state
	if params.Paused {
		transfer.Pause()
	}
	
	// Add to maps
	s.transfers[hashStr] = transfer
	handle := NewTransferHandle(transfer, s)
	s.transferHandles[hashStr] = handle
	
	// Save initial resume data
	if err := s.persistenceManager.SaveTransfer(transfer); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to save initial resume data: %v\n", err)
	}
	
	return handle, nil
}

// AddTransferFromLink adds a transfer from an ed2k:// link
func (s *Session) AddTransferFromLink(linkStr, downloadDir string) (*TransferHandle, error) {
	link, err := ParseEMuleLink(linkStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse link: %v", err)
	}
	
	if !link.IsFileLink() {
		return nil, fmt.Errorf("only file links are supported for transfers")
	}
	
	// Use default download directory if not specified
	if downloadDir == "" {
		downloadDir = s.settings.IncomingDirectory
	}
	
	params, err := NewAddTransferParams(link, downloadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer params: %v", err)
	}
	
	return s.AddTransfer(params)
}

// RemoveTransfer removes a transfer from the session
func (s *Session) RemoveTransfer(transferHash *hash.Hash, deleteFiles bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	hashStr := transferHash.String()
	
	transferToRemove, exists := s.transfers[hashStr]
	if !exists {
		return fmt.Errorf("transfer not found")
	}
	
	// Remove from maps
	delete(s.transfers, hashStr)
	delete(s.transferHandles, hashStr)
	
	// Remove resume data
	s.persistenceManager.RemoveTransfer(transferHash)
	
	// Delete files if requested
	if deleteFiles {
		// TODO: Delete download files
		_ = transferToRemove // Use the variable to avoid unused error
	}
	
	return nil
}

// GetTransfers returns all transfer handles
func (s *Session) GetTransfers() []*TransferHandle {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	handles := make([]*TransferHandle, 0, len(s.transferHandles))
	for _, handle := range s.transferHandles {
		handles = append(handles, handle)
	}
	
	return handles
}

// GetTransfer returns a specific transfer handle
func (s *Session) GetTransfer(transferHash *hash.Hash) *TransferHandle {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	hashStr := transferHash.String()
	return s.transferHandles[hashStr]
}

// PauseAll pauses all transfers
func (s *Session) PauseAll() error {
	handles := s.GetTransfers()
	
	for _, handle := range handles {
		if err := handle.Pause(); err != nil {
			return fmt.Errorf("failed to pause transfer %s: %v", handle.GetHash().String(), err)
		}
	}
	
	return nil
}

// ResumeAll resumes all transfers
func (s *Session) ResumeAll() error {
	handles := s.GetTransfers()
	
	for _, handle := range handles {
		if err := handle.Resume(); err != nil {
			return fmt.Errorf("failed to resume transfer %s: %v", handle.GetHash().String(), err)
		}
	}
	
	return nil
}

// GetSessionStats returns overall session statistics
func (s *Session) GetSessionStats() SessionStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := SessionStats{
		TotalTransfers:      len(s.transfers),
		ActiveTransfers:     0,
		PausedTransfers:     0,
		CompletedTransfers:  0,
		TotalDownloaded:     0,
		TotalUploaded:       0,
		GlobalDownloadRate:  0,
		GlobalUploadRate:    0,
		ConnectedPeers:      0,
		KnownServers:        len(s.serverList.GetServers()),
		KnownNodes:          len(s.nodesData.GetNodes()),
	}
	
	for _, transfer := range s.transfers {
		status := transfer.GetStatus()
		
		switch status.State {
		case TransferStateDownloading:
			stats.ActiveTransfers++
		case TransferStatePaused:
			stats.PausedTransfers++
		case TransferStateCompleted:
			stats.CompletedTransfers++
		}
		
		stats.TotalDownloaded += status.Downloaded
		stats.TotalUploaded += status.Uploaded
		stats.GlobalDownloadRate += status.DownloadRate
		stats.GlobalUploadRate += status.UploadRate
		stats.ConnectedPeers += status.ConnectedPeers
	}
	
	return stats
}

// mainLoop is the main session loop
func (s *Session) mainLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.updateTransfers()
		}
	}
}

// updateTransfers updates all transfer statistics
func (s *Session) updateTransfers() {
	s.mutex.RLock()
	transfers := make([]*Transfer, 0, len(s.transfers))
	for _, transfer := range s.transfers {
		transfers = append(transfers, transfer)
	}
	s.mutex.RUnlock()
	
	// Update each transfer (simulation)
	for _, transfer := range transfers {
		if !transfer.IsPaused() {
			status := transfer.GetStatus()
			
			// Start queued transfers
			if status.State == TransferStateQueued {
				// Simulate connection to peers and start downloading
				s.simulateTransferStart(transfer)
			} else if status.State == TransferStateDownloading && status.Downloaded < status.Size {
				// Continue download simulation
				s.simulateDownloadProgress(transfer)
			}
		}
	}
}

// simulateTransferStart simulates starting a queued transfer
func (s *Session) simulateTransferStart(transfer *Transfer) {
	// Simulate finding peers and starting download
	simulatedRate := float64(50*1024 + (time.Now().UnixNano()%100)*1024) // 50-150 KB/s
	
	// Add some simulated peers
	s.simulatePeers(transfer)
	
	// Start with current downloaded amount and the simulated rate
	status := transfer.GetStatus()
	transfer.updateStats(status.Downloaded, status.Uploaded, simulatedRate, 0)
}

// simulateDownloadProgress simulates ongoing download progress
func (s *Session) simulateDownloadProgress(transfer *Transfer) {
	status := transfer.GetStatus()
	
	// Simulate variable download rates (realistic behavior)
	baseRate := status.DownloadRate
	variation := 0.8 + (float64(time.Now().UnixNano()%400) / 1000.0) // 0.8 to 1.2 multiplier
	newRate := baseRate * variation
	
	// Simulate downloaded bytes based on rate (1 second interval)
	bytesDownloaded := int64(newRate)
	newDownloaded := status.Downloaded + bytesDownloaded
	if newDownloaded > status.Size {
		newDownloaded = status.Size
	}
	
	// Simulate some upload as well
	newUploaded := status.Uploaded + int64(newRate*0.1) // 10% of download rate
	
	transfer.updateStats(newDownloaded, newUploaded, newRate, newRate*0.1)
}

// simulatePeers adds simulated peer connections to a transfer
func (s *Session) simulatePeers(transfer *Transfer) {
	// Add 2-5 simulated peers
	peerCount := 2 + (time.Now().UnixNano() % 4)
	
	for i := int64(0); i < peerCount; i++ {
		// Create IP address as uint32 (192.168.1.100 + i)
		ip := uint32(0xC0A80164) + uint32(i) // 192.168.1.100 in network byte order
		port := uint16(4661 + i)
		
		peer := &PeerInfo{
			Endpoint:     protocol.NewEndpointFromIPPort(ip, port),
			UserHash:     hash.NewHash(),
			ClientName:   fmt.Sprintf("eMule %d.%d.%d", 0, 50, i),
			Downloaded:   0,
			Uploaded:     0,
			DownloadRate: float64(10*1024 + (i*5*1024)), // 10-30 KB/s per peer
			UploadRate:   0,
			Connected:    true,
		}
		transfer.addPeer(peer)
	}
}

// SessionStats contains overall session statistics
type SessionStats struct {
	TotalTransfers      int     `json:"total_transfers"`
	ActiveTransfers     int     `json:"active_transfers"`
	PausedTransfers     int     `json:"paused_transfers"`
	CompletedTransfers  int     `json:"completed_transfers"`
	TotalDownloaded     int64   `json:"total_downloaded"`
	TotalUploaded       int64   `json:"total_uploaded"`
	GlobalDownloadRate  float64 `json:"global_download_rate"`
	GlobalUploadRate    float64 `json:"global_upload_rate"`
	ConnectedPeers      int     `json:"connected_peers"`
	KnownServers        int     `json:"known_servers"`
	KnownNodes          int     `json:"known_nodes"`
}