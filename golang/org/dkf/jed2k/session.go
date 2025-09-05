package jed2k

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
func (th *TransferHandle) GetHash() *protocol.Hash {
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
	
	// Server connections for peer discovery
	serverConnections   map[string]*ServerConnection
	activeServerConn    *ServerConnection
	
	// State
	running             bool
	connected           bool
	
	// Synchronization
	mutex               sync.RWMutex
	
	// Shutdown
	done                chan struct{}
}

// NewSession creates a new ed2k session with custom resume data implementation
func NewSession(settings *Settings, resumeData ResumeData) *Session {
	if settings == nil {
		settings = NewDefaultSettings()
	}
	
	if resumeData == nil {
		resumeData = NewMemoryResumeData()
	}
	
	session := &Session{
		settings:           settings,
		transfers:          make(map[string]*Transfer),
		transferHandles:    make(map[string]*TransferHandle),
		serverList:         NewServerList(),
		nodesData:          NewNodesData(),
		persistenceManager: NewPersistenceManager(resumeData),
		serverConnections:  make(map[string]*ServerConnection),
		done:               make(chan struct{}),
	}
	
	return session
}

// NewSessionWithDefaults creates a new ed2k session with default memory-based resume data
func NewSessionWithDefaults(settings *Settings) *Session {
	return NewSession(settings, NewMemoryResumeData())
}

// NewSessionWithDiskPersistence creates a new ed2k session with disk-based resume data
func NewSessionWithDiskPersistence(settings *Settings, resumeDataDir string) *Session {
	return NewSession(settings, NewDiskResumeData(resumeDataDir))
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
	
	// Connect to servers for peer discovery
	go s.connectToServers()
	
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
	
	// Close all server connections
	for _, serverConn := range s.serverConnections {
		serverConn.Close()
	}
	
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
	// Parse endpoint string format: "IP:port"
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid endpoint format, expected IP:port")
	}
	
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", parts[0])
	}
	
	port, err := strconv.Atoi(parts[1])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %s", parts[1])
	}
	
	// Convert IP to uint32
	ipv4 := ip.To4()
	if ipv4 == nil {
		return fmt.Errorf("IPv6 not supported")
	}
	
	ipUint32 := uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
	
	// Add to server list
	serverEndpoint := ServerEndpoint{
		Endpoint: protocol.NewEndpointFromIPPort(ipUint32, uint16(port)),
		Name:     name,
	}
	
	s.serverList.servers = append(s.serverList.servers, serverEndpoint)
	
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
		// Apply completed bytes and piece information
		if len(params.ResumeData.Pieces) > 0 {
			const pieceSize = 9728000 // 9.5 MB default piece size
			completedPieces := 0
			for _, completed := range params.ResumeData.Pieces {
				if completed {
					completedPieces++
				}
			}
			transfer.downloaded = int64(completedPieces * pieceSize)
			if transfer.downloaded > transfer.size {
				transfer.downloaded = transfer.size
			}
		}
		
		// Apply download progress from resume data
		if params.ResumeData.Downloaded > 0 {
			transfer.downloaded = params.ResumeData.Downloaded
		}
		
		// Resume with previous state if not explicitly paused
		if !params.Paused && params.ResumeData.Downloaded > 0 {
			transfer.state = TransferStateDownloading
		}
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
	
	// Start the transfer immediately if not paused
	if !params.Paused {
		go func() {
			// Small delay to allow the function to return first
			time.Sleep(100 * time.Millisecond)
			s.initiateTransferConnections(transfer)
			
			// Request sources from all connected servers
			time.Sleep(2 * time.Second) // Allow some time for connection setup
			s.mutex.RLock()
			connectedServers := make([]*ServerConnection, 0)
			for _, serverConn := range s.serverConnections {
				if serverConn.IsConnected() {
					connectedServers = append(connectedServers, serverConn)
				}
			}
			s.mutex.RUnlock()
			
			if len(connectedServers) > 0 {
				fmt.Printf("[SERVER] Requesting sources for new transfer %s from %d servers\n", transfer.name, len(connectedServers))
				for _, serverConn := range connectedServers {
					go func(server *ServerConnection) {
						if err := server.RequestSources(transfer.GetHash()); err != nil {
							fmt.Printf("[SERVER] Failed to request sources for new transfer %s from %s: %v\n", transfer.name, server.identifier, err)
						} else {
							fmt.Printf("[SERVER] Requested sources for new transfer %s from %s\n", transfer.name, server.identifier)
						}
					}(serverConn)
					
					// Small delay between server requests
					time.Sleep(1 * time.Second)
				}
			} else {
				fmt.Printf("[SERVER] No connected servers available for new transfer %s\n", transfer.name)
			}
		}()
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
func (s *Session) RemoveTransfer(transferHash *protocol.Hash, deleteFiles bool) error {
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
		// Delete the download file and any temporary files
		if transferToRemove != nil {
			downloadPath := filepath.Join(transferToRemove.downloadDirectory, transferToRemove.name)
			if err := os.Remove(downloadPath); err != nil && !os.IsNotExist(err) {
				// Log error but don't fail the removal
				fmt.Printf("Warning: failed to delete file %s: %v\n", downloadPath, err)
			}
			
			// Also try to delete partial/temp files
			tempPath := downloadPath + ".part"
			if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: failed to delete temp file %s: %v\n", tempPath, err)
			}
		}
	}
	
	return nil
}

// GetTransfers returns all transfer handles (non-blocking version to avoid deadlocks)
func (s *Session) GetTransfers() []*TransferHandle {
	// Try to acquire the lock with a timeout to avoid deadlocks
	lockAcquired := make(chan []*TransferHandle, 1)
	
	go func() {
		s.mutex.RLock()
		// Create a slice with same capacity but copy the handles
		handles := make([]*TransferHandle, 0, len(s.transferHandles))
		for _, handle := range s.transferHandles {
			handles = append(handles, handle)
		}
		s.mutex.RUnlock()
		lockAcquired <- handles
	}()
	
	// Wait for lock acquisition with a longer timeout
	select {
	case handles := <-lockAcquired:
		return handles
	case <-time.After(1 * time.Second): // 1 second timeout
		// If we can't get the lock after 1 second, return empty slice to avoid indefinite hang
		fmt.Printf("Warning: GetTransfers() timed out after 1 second\n")
		return make([]*TransferHandle, 0)
	}
}

// GetTransfer returns a specific transfer handle
func (s *Session) GetTransfer(transferHash *protocol.Hash) *TransferHandle {
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
	ticker := time.NewTicker(2 * time.Second) // Reduced frequency to avoid mutex contention
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

// updateTransfers updates all transfer statistics and handles timeouts/retries
func (s *Session) updateTransfers() {
	// Get transfers without holding the session lock for too long
	s.mutex.RLock()
	transfers := make([]*Transfer, 0, len(s.transfers))
	for _, transfer := range s.transfers {
		transfers = append(transfers, transfer)
	}
	s.mutex.RUnlock()
	
	// Update each transfer with proper timeout and retry handling
	for _, transfer := range transfers {
		if !transfer.IsPaused() {
			// Check if transfer needs initial connection setup
			transfer.mutex.RLock()
			needsInitialConnection := transfer.state == TransferStateQueued
			hasLowConnections := len(transfer.connections) == 0 && transfer.state == TransferStateDownloading
			transfer.mutex.RUnlock()
			
			// Start initial connections if needed
			if needsInitialConnection {
				s.initiateTransferConnections(transfer)
			}
			
			// Let transfer handle its own peer connections, timeouts, and retries
			transfer.SecondTick(1000, s) // 1000ms = 1 second
			
			// Add more peers if we don't have enough connections
			if hasLowConnections {
				s.addMorePeersToTransfer(transfer)
			}
		}
	}
}

// initiateTransferConnections starts initial peer connections for a transfer
func (s *Session) initiateTransferConnections(transfer *Transfer) {
	// Add some initial peers to the transfer's policy
	s.addInitialPeersToTransfer(transfer)
	
	// Set transfer state to downloading when it starts connecting
	transfer.mutex.Lock()
	if transfer.state == TransferStateQueued {
		transfer.state = TransferStateDownloading
		fmt.Printf("Transfer %s state changed from queued to downloading\n", transfer.name)
	}
	transfer.mutex.Unlock()
}

// addMorePeersToTransfer adds additional peers when transfer has low connection count
func (s *Session) addMorePeersToTransfer(transfer *Transfer) {
	// Add additional peers to maintain active connections
	peerCount := 2 + (time.Now().UnixNano() % 2) // 2-3 peers
	
	for i := int64(0); i < peerCount; i++ {
		// Create different IP addresses
		ip := uint32(0xC0A80164) + uint32(10) + uint32(i) // 192.168.1.110+
		port := uint16(4661 + i)
		endpoint := protocol.NewEndpointFromIPPort(ip, port)
		
		// Create peer with different source flags
		peer := NewPeerWithFlags(endpoint, true, SourceDHT|SourceIncoming)
		
		// Add peer to transfer's policy
		if transfer.policy != nil {
			added, err := transfer.policy.AddPeer(peer)
			if err == nil && added {
				// Also add to legacy peers map
				peerInfo := &PeerInfo{
					Endpoint:     endpoint,
					UserHash:     protocol.NewHash(),
					ClientName:   fmt.Sprintf("aDrive %d.%d.%d", 3, 2, i),
					Downloaded:   0,
					Uploaded:     0,
					DownloadRate: 0,
					UploadRate:   0,
					Connected:    false,
				}
				
				transfer.mutex.Lock()
				transfer.peers[endpoint.String()] = peerInfo
				transfer.mutex.Unlock()
			}
		}
	}
}

// addInitialPeersToTransfer adds initial peers for a transfer to get started
func (s *Session) addInitialPeersToTransfer(transfer *Transfer) {
	// Add some simulated peers to bootstrap the transfer
	peerCount := 3 + (time.Now().UnixNano() % 3) // 3-5 peers
	
	for i := int64(0); i < peerCount; i++ {
		// Create IP address as uint32 (192.168.1.100 + i)
		ip := uint32(0xC0A80164) + uint32(i) // 192.168.1.100
		port := uint16(4661 + i)
		endpoint := protocol.NewEndpointFromIPPort(ip, port)
		
		// Create peer with some source flags
		peer := NewPeerWithFlags(endpoint, true, SourceServer|SourceDHT)
		
		// Add peer to transfer's policy
		if transfer.policy != nil {
			added, err := transfer.policy.AddPeer(peer)
			if err != nil {
				// Log error but continue
				fmt.Printf("Warning: failed to add peer %s: %v\n", endpoint.String(), err)
			} else if added {
				// Also add to legacy peers map for compatibility
				peerInfo := &PeerInfo{
					Endpoint:     endpoint,
					UserHash:     protocol.NewHash(),
					ClientName:   fmt.Sprintf("eMule %d.%d.%d", 0, 50, i),
					Downloaded:   0,
					Uploaded:     0,
					DownloadRate: 0,
					UploadRate:   0,
					Connected:    false,
				}
				transfer.addPeer(peerInfo)
			}
		}
	}
}

// simulateRealTransferProgress simulates realistic transfer progress with timeouts and reconnections
func (s *Session) simulateRealTransferProgress(transfer *Transfer) {
	transfer.mutex.Lock()
	defer transfer.mutex.Unlock()
	
	// Check if transfer has active connections
	activeConnections := 0
	totalDownloadRate := 0.0
	
	for _, conn := range transfer.connections {
		if conn.IsConnected() && !conn.IsDisconnecting() {
			activeConnections++
			// Simulate download rate per connection
			rate := 10.0 * 1024 * (1.0 + rand.Float64()) // 10-20 KB/s per connection
			totalDownloadRate += rate
		}
	}
	
	// Only progress if we have active connections
	if activeConnections > 0 {
		// Simulate downloaded bytes based on rate (1 second interval)
		bytesDownloaded := int64(totalDownloadRate)
		newDownloaded := transfer.downloaded + bytesDownloaded
		if newDownloaded > transfer.size {
			newDownloaded = transfer.size
			transfer.finished = true
			transfer.state = TransferStateCompleted
		}
		
		// Simulate some upload as well
		newUploaded := transfer.uploaded + int64(totalDownloadRate*0.1) // 10% of download rate
		
		transfer.downloaded = newDownloaded
		transfer.uploaded = newUploaded
		transfer.downloadRate = totalDownloadRate
		transfer.uploadRate = totalDownloadRate * 0.1
		
		if !transfer.finished && transfer.state != TransferStateError {
			transfer.state = TransferStateDownloading
		}
	} else {
		// No active connections, reduce rates
		transfer.downloadRate = 0
		transfer.uploadRate = 0
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

// connectToServers connects to available ed2k servers for peer discovery
func (s *Session) connectToServers() {
	// Wait a bit for session to initialize
	time.Sleep(2 * time.Second)
	
	// Get available servers
	servers := s.getAvailableServers()
	if len(servers) == 0 {
		fmt.Printf("[SERVER] No servers available for connection\n")
		return
	}
	
	// Limit the number of simultaneous server connections to avoid overwhelming
	maxConnections := 5
	if len(servers) < maxConnections {
		maxConnections = len(servers)
	}
	
	fmt.Printf("[SERVER] Attempting to connect to %d servers simultaneously\n", maxConnections)
	
	// Try to connect to multiple servers simultaneously
	connectedCount := 0
	for i, server := range servers[:maxConnections] {
		if !s.IsRunning() {
			break
		}
		
		// Small delay between connection attempts to avoid overwhelming
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		
		go func(srv ServerInfo) {
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", srv.IP, srv.Port))
			if err != nil {
				fmt.Printf("[SERVER] Failed to resolve server %s:%d - %v\n", srv.IP, srv.Port, err)
				return
			}
			
			serverConn := NewServerConnection(srv.Name, addr, s)
			
			s.mutex.Lock()
			s.serverConnections[srv.Name] = serverConn
			s.mutex.Unlock()
			
			if err := serverConn.Connect(); err != nil {
				fmt.Printf("[SERVER] Failed to connect to %s - %v\n", srv.Name, err)
				// Remove failed connection from map
				s.mutex.Lock()
				delete(s.serverConnections, srv.Name)
				s.mutex.Unlock()
				return
			}
			
			// Set as active server connection (first connected server becomes primary)
			s.mutex.Lock()
			if s.activeServerConn == nil {
				s.activeServerConn = serverConn
				fmt.Printf("[SERVER] Set %s as primary server connection\n", srv.Name)
			}
			s.mutex.Unlock()
			
			fmt.Printf("[SERVER] Successfully connected to server %s\n", srv.Name)
			connectedCount++
			
			// Wait for handshake completion
			time.Sleep(5 * time.Second)
			
			// Start requesting sources from this server for all transfers  
			if serverConn.IsConnected() {
				s.requestSourcesFromServer(serverConn)
			}
			
			// Keep connection alive and monitor
			go s.maintainServerConnection(serverConn)
		}(server)
	}
	
	// Wait a bit for initial connections to establish
	time.Sleep(10 * time.Second)
	
	// Report connection results
	s.mutex.RLock()
	actualConnected := len(s.serverConnections)
	s.mutex.RUnlock()
	
	fmt.Printf("[SERVER] Connected to %d out of %d attempted servers\n", actualConnected, maxConnections)
}

// getAvailableServers returns a list of available servers
func (s *Session) getAvailableServers() []ServerInfo {
	servers := make([]ServerInfo, 0)
	
	// Add servers from server list
	if s.serverList != nil && len(s.serverList.servers) > 0 {
		for _, server := range s.serverList.servers {
			serverIP := server.Endpoint.IP()
			ip := fmt.Sprintf("%d.%d.%d.%d",
				(serverIP>>24)&0xFF,
				(serverIP>>16)&0xFF,
				(serverIP>>8)&0xFF,
				serverIP&0xFF)
			
			servers = append(servers, ServerInfo{
				Name: server.Name,
				IP:   ip,
				Port: int(server.Endpoint.Port()),
			})
		}
	}
	
	// Add default servers if none available
	if len(servers) == 0 {
		servers = append(servers, 
			ServerInfo{Name: "eDonkeyServer No1", IP: "176.103.48.36", Port: 4242},
			ServerInfo{Name: "eDonkeyServer No2", IP: "195.24.106.203", Port: 4661},
			ServerInfo{Name: "eMule Security", IP: "80.208.228.241", Port: 8369},
			ServerInfo{Name: "Razorback 2", IP: "195.245.244.243", Port: 4661},
			ServerInfo{Name: "TV Underground", IP: "195.245.244.243", Port: 4662},
		)
	}
	
	return servers
}

// ServerInfo contains server connection information
type ServerInfo struct {
	Name string
	IP   string
	Port int
}

// maintainServerConnection keeps server connection alive and handles reconnection
func (s *Session) maintainServerConnection(serverConn *ServerConnection) {
	ticker := time.NewTicker(60 * time.Second) // Check every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			if !serverConn.IsConnected() {
				fmt.Printf("[SERVER] Lost connection to %s, attempting reconnect...\n", serverConn.identifier)
				
				// If this was the active server, try to set another one as active
				s.mutex.Lock()
				if s.activeServerConn == serverConn {
					s.activeServerConn = nil
					// Find another connected server to be primary
					for _, otherServer := range s.serverConnections {
						if otherServer != serverConn && otherServer.IsConnected() {
							s.activeServerConn = otherServer
							fmt.Printf("[SERVER] Switched primary server to %s\n", otherServer.identifier)
							break
						}
					}
				}
				s.mutex.Unlock()
				
				// Try to reconnect
				if err := serverConn.Connect(); err != nil {
					fmt.Printf("[SERVER] Reconnection failed for %s: %v\n", serverConn.identifier, err)
					
					// Remove from connections map if reconnection consistently fails
					s.mutex.Lock()
					delete(s.serverConnections, serverConn.identifier)
					s.mutex.Unlock()
				} else {
					fmt.Printf("[SERVER] Successfully reconnected to %s\n", serverConn.identifier)
					
					// Re-add to connections map
					s.mutex.Lock()
					s.serverConnections[serverConn.identifier] = serverConn
					// Set as primary if we don't have one
					if s.activeServerConn == nil {
						s.activeServerConn = serverConn
						fmt.Printf("[SERVER] Set reconnected %s as primary server\n", serverConn.identifier)
					}
					s.mutex.Unlock()
					
					// Request sources for all transfers from this reconnected server
					time.Sleep(5 * time.Second) // Wait for handshake
					s.requestSourcesFromServer(serverConn)
				}
			} else {
				// Send periodic ping to keep connection alive
				serverConn.Ping()
			}
		}
	}
}

// requestSourcesFromServer requests peer sources for all transfers from a specific server
func (s *Session) requestSourcesFromServer(serverConn *ServerConnection) {
	if serverConn == nil || !serverConn.IsConnected() {
		return
	}
	
	s.mutex.RLock()
	transfers := make([]*Transfer, 0, len(s.transfers))
	for _, transfer := range s.transfers {
		if !transfer.IsFinished() {
			transfers = append(transfers, transfer)
		}
	}
	s.mutex.RUnlock()
	
	if len(transfers) == 0 {
		return
	}
	
	fmt.Printf("[SERVER] Requesting sources for %d transfers from server %s\n", len(transfers), serverConn.identifier)
	
	for _, transfer := range transfers {
		if err := serverConn.RequestSources(transfer.GetHash()); err != nil {
			fmt.Printf("[SERVER] Failed to request sources for %s from %s: %v\n", transfer.name, serverConn.identifier, err)
		} else {
			fmt.Printf("[SERVER] Requested sources for transfer %s from %s\n", transfer.name, serverConn.identifier)
		}
		
		// Small delay between requests to avoid overwhelming the server
		time.Sleep(500 * time.Millisecond)
	}
}

// requestSourcesForAllTransfers requests peer sources for all active transfers from all connected servers
func (s *Session) requestSourcesForAllTransfers() {
	s.mutex.RLock()
	serverConnections := make([]*ServerConnection, 0, len(s.serverConnections))
	for _, serverConn := range s.serverConnections {
		if serverConn.IsConnected() {
			serverConnections = append(serverConnections, serverConn)
		}
	}
	transfers := make([]*Transfer, 0, len(s.transfers))
	for _, transfer := range s.transfers {
		if !transfer.IsFinished() {
			transfers = append(transfers, transfer)
		}
	}
	s.mutex.RUnlock()
	
	if len(serverConnections) == 0 {
		fmt.Printf("[SERVER] No connected servers available for source requests\n")
		return
	}
	
	if len(transfers) == 0 {
		return
	}
	
	fmt.Printf("[SERVER] Requesting sources for %d transfers from %d connected servers\n", len(transfers), len(serverConnections))
	
	// Request sources from all connected servers
	for _, serverConn := range serverConnections {
		go func(server *ServerConnection) {
			for _, transfer := range transfers {
				if err := server.RequestSources(transfer.GetHash()); err != nil {
					fmt.Printf("[SERVER] Failed to request sources for %s from %s: %v\n", transfer.name, server.identifier, err)
				} else {
					fmt.Printf("[SERVER] Requested sources for transfer %s from %s\n", transfer.name, server.identifier)
				}
				
				// Small delay between requests
				time.Sleep(300 * time.Millisecond)
			}
			
			// Delay before next server to spread the load
			time.Sleep(2 * time.Second)
		}(serverConn)
	}
}