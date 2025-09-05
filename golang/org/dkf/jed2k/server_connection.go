package jed2k

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// ServerConnection handles communication with ed2k servers
type ServerConnection struct {
	conn           net.Conn
	session        *Session
	identifier     string
	address        *net.TCPAddr
	handshakeDone  bool
	lastPingTime   int64
	mutex          sync.RWMutex
	connected      bool
	clientId       uint32
	debugLogging   bool
}

// NewServerConnection creates a new server connection
func NewServerConnection(identifier string, address *net.TCPAddr, session *Session) *ServerConnection {
	return &ServerConnection{
		session:      session,
		identifier:   identifier,
		address:      address,
		debugLogging: true, // Enable detailed protocol logging
	}
}

// Connect establishes connection to ed2k server
func (sc *ServerConnection) Connect() error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if sc.connected {
		return nil
	}

	if sc.debugLogging {
		fmt.Printf("[SERVER] Connecting to server %s (%s)\n", sc.identifier, sc.address.String())
	}

	conn, err := net.DialTCP("tcp", nil, sc.address)
	if err != nil {
		if sc.debugLogging {
			fmt.Printf("[SERVER] Failed to connect to %s: %v\n", sc.identifier, err)
		}
		return err
	}

	sc.conn = conn
	sc.connected = true
	sc.lastPingTime = time.Now().Unix()

	if sc.debugLogging {
		fmt.Printf("[SERVER] Connected to server %s successfully\n", sc.identifier)
	}

	// Start handshake immediately after connection
	go sc.performHandshake()
	go sc.readLoop()

	return nil
}

// performHandshake performs the ed2k server login handshake
func (sc *ServerConnection) performHandshake() {
	if sc.debugLogging {
		fmt.Printf("[SERVER] Starting handshake with %s\n", sc.identifier)
	}

	// Create login request packet
	loginPacket := sc.createLoginRequest()
	if err := sc.sendPacket(0xE3, 0x01, loginPacket); err != nil {
		if sc.debugLogging {
			fmt.Printf("[SERVER] Failed to send login request to %s: %v\n", sc.identifier, err)
		}
		sc.Close()
		return
	}

	if sc.debugLogging {
		fmt.Printf("[SERVER] → Login request sent to %s\n", sc.identifier)
	}
}

// createLoginRequest creates a login request packet for server handshake
func (sc *ServerConnection) createLoginRequest() []byte {
	// Basic login request structure for ed2k protocol
	packet := make([]byte, 0, 256)
	
	// User hash (16 bytes)
	userHash := sc.session.settings.UserAgent.Bytes()
	packet = append(packet, userHash[:]...)
	
	// Client endpoint (6 bytes: 4 bytes IP + 2 bytes port)
	ip := uint32(0) // Server will provide our IP
	port := uint16(sc.session.settings.ListenPort)
	
	ipBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(ipBytes, ip)
	packet = append(packet, ipBytes...)
	
	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, port)
	packet = append(packet, portBytes...)
	
	// Tags (capabilities and client info)
	tags := sc.createClientTags()
	packet = append(packet, tags...)
	
	if sc.debugLogging {
		fmt.Printf("[SERVER] Created login request: %d bytes\n", len(packet))
	}
	
	return packet
}

// createClientTags creates capability tags for login
func (sc *ServerConnection) createClientTags() []byte {
	tags := make([]byte, 0, 64)
	
	// Number of tags
	tags = append(tags, 4) // 4 tags
	
	// CT_VERSION tag
	tags = append(tags, 0x01, 0x03, 0x3c, 0x00, 0x00, 0x00) // Version = 0x3c
	
	// CT_SERVER_FLAGS tag  
	serverFlags := uint32(0x0f) // Basic capabilities
	tags = append(tags, 0x20, 0x03)
	flagBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(flagBytes, serverFlags)
	tags = append(tags, flagBytes...)
	
	// CT_NAME tag
	clientName := "ged2k Go Client"
	tags = append(tags, 0x01, 0x02)
	tags = append(tags, uint8(len(clientName)))
	tags = append(tags, []byte(clientName)...)
	
	// CT_EMULE_VERSION tag
	emuleVersion := uint32(0x30000000) // Basic version
	tags = append(tags, 0x14, 0x03)
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, emuleVersion)
	tags = append(tags, versionBytes...)
	
	return tags
}

// readLoop continuously reads packets from server
func (sc *ServerConnection) readLoop() {
	defer sc.Close()
	
	buffer := make([]byte, 4096)
	
	for {
		sc.mutex.RLock()
		connected := sc.connected
		conn := sc.conn
		sc.mutex.RUnlock()
		
		if !connected || conn == nil {
			break
		}
		
		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF && sc.debugLogging {
				fmt.Printf("[SERVER] Read error from %s: %v\n", sc.identifier, err)
			}
			break
		}
		
		if n > 0 {
			sc.processIncomingData(buffer[:n])
		}
	}
}

// processIncomingData processes incoming server packets
func (sc *ServerConnection) processIncomingData(data []byte) {
	if len(data) < 6 {
		return // Too short for valid packet
	}
	
	// Parse ed2k packet header: [protocol][size][opcode]
	protocol := data[0]
	size := binary.LittleEndian.Uint32(data[1:5])
	opcode := data[5]
	
	if sc.debugLogging {
		fmt.Printf("[SERVER] ← Received packet from %s: protocol=0x%02x size=%d opcode=0x%02x\n", 
			sc.identifier, protocol, size, opcode)
	}
	
	if len(data) < int(size)+5 {
		if sc.debugLogging {
			fmt.Printf("[SERVER] Incomplete packet received, waiting for more data\n")
		}
		return
	}
	
	payload := data[6:6+size-1] // Exclude opcode from payload
	
	switch opcode {
	case 0x40: // Server status
		sc.handleServerStatus(payload)
	case 0x41: // Server ID
		sc.handleServerID(payload)
	case 0x42: // Server message
		sc.handleServerMessage(payload)
	case 0x43: // ID change
		sc.handleIDChange(payload)
	case 0x44: // Server info
		sc.handleServerInfo(payload)  
	case 0x46: // Found file sources
		sc.handleFoundFileSources(payload)
	default:
		if sc.debugLogging {
			fmt.Printf("[SERVER] Unknown opcode 0x%02x from %s\n", opcode, sc.identifier)
		}
	}
}

// handleServerID handles server ID assignment
func (sc *ServerConnection) handleServerID(payload []byte) {
	if len(payload) >= 4 {
		sc.clientId = binary.LittleEndian.Uint32(payload[:4])
		sc.handshakeDone = true
		
		if sc.debugLogging {
			fmt.Printf("[SERVER] ← Server ID received: %d (0x%08x) from %s\n", 
				sc.clientId, sc.clientId, sc.identifier)
		}
		
		// Handshake complete, we can now start requesting file sources
		if sc.debugLogging {
			fmt.Printf("[SERVER] Handshake completed with %s, ready for file requests\n", sc.identifier)
		}
	}
}

// handleFoundFileSources handles peer source responses from server
func (sc *ServerConnection) handleFoundFileSources(payload []byte) {
	if len(payload) < 16 {
		return // Need at least hash
	}
	
	// Extract file hash
	var hash protocol.Hash
	copy(hash.Bytes(), payload[:16])
	
	// Parse peer count
	if len(payload) < 17 {
		return
	}
	peerCount := payload[16]
	
	// Always log peer discovery (not just debug mode)
	fmt.Printf("[PEER DISCOVERY] Server %s returned %d peer sources for hash %s\n", 
		sc.identifier, peerCount, hash.String()[:8]+"...")
	
	// Parse peer endpoints
	offset := 17
	peers := make([]*protocol.Endpoint, 0, peerCount)
	
	for i := 0; i < int(peerCount) && offset+6 <= len(payload); i++ {
		ip := binary.LittleEndian.Uint32(payload[offset:offset+4])
		port := binary.LittleEndian.Uint16(payload[offset+4:offset+6])
		endpoint := protocol.NewEndpointFromIPPort(ip, port)
		peers = append(peers, endpoint)
		offset += 6
		
		// Always log each peer discovered
		fmt.Printf("[PEER DISCOVERY]   Peer %d: %s (from server %s)\n", 
			i+1, endpoint.String(), sc.identifier)
	}
	
	// Add peers to corresponding transfer
	sc.addPeersToTransfer(&hash, peers)
}

// addPeersToTransfer adds discovered peers to a transfer
func (sc *ServerConnection) addPeersToTransfer(hash *protocol.Hash, peers []*protocol.Endpoint) {
	// Find transfer by hash
	sc.session.mutex.RLock()
	transfer, exists := sc.session.transfers[hash.String()]
	sc.session.mutex.RUnlock()
	
	if !exists {
		if sc.debugLogging {
			fmt.Printf("[SERVER] No transfer found for hash %s\n", hash.String())
		}
		return
	}
	
	// Add each peer to the transfer's policy
	addedCount := 0
	for _, endpoint := range peers {
		peer := NewPeerWithFlags(endpoint, true, SourceServer)
		if transfer.policy != nil {
			added, err := transfer.policy.AddPeer(peer)
			if err == nil && added {
				addedCount++
				
				// Also add to legacy peers info
				peerInfo := &PeerInfo{
					Endpoint:     endpoint,
					UserHash:     protocol.NewHash(),
					ClientName:   "Server Peer",
					Downloaded:   0,
					Uploaded:     0,
					DownloadRate: 0,
					UploadRate:   0,
					Connected:    false,
				}
				
				transfer.mutex.Lock()
				transfer.peers[endpoint.String()] = peerInfo
				transfer.mutex.Unlock()
				
				fmt.Printf("[PEER ADDED] %s added to transfer %s (source: server %s)\n", 
					endpoint.String(), transfer.name, sc.identifier)
			} else if err != nil {
				fmt.Printf("[PEER REJECTED] %s rejected for transfer %s: %v\n", 
					endpoint.String(), transfer.name, err)
			}
		}
	}
	
	fmt.Printf("[PEER SUMMARY] Added %d/%d new peers to transfer %s from server %s\n", 
		addedCount, len(peers), transfer.name, sc.identifier)
}

// RequestSources requests peer sources for a file hash from server
func (sc *ServerConnection) RequestSources(hash *protocol.Hash) error {
	sc.mutex.RLock()
	connected := sc.connected && sc.handshakeDone
	sc.mutex.RUnlock()
	
	if !connected {
		return fmt.Errorf("server not connected or handshake not complete")
	}
	
	if sc.debugLogging {
		fmt.Printf("[SERVER] → Requesting sources for hash %s from %s\n", hash.String(), sc.identifier)
	}
	
	// Create get sources request packet
	packet := make([]byte, 16)
	copy(packet, hash.Bytes())
	
	return sc.sendPacket(0xE3, 0x19, packet) // GetSources opcode
}

// sendPacket sends a packet to the server with proper ed2k framing
func (sc *ServerConnection) sendPacket(protocol, opcode byte, payload []byte) error {
	sc.mutex.RLock()
	conn := sc.conn
	connected := sc.connected
	sc.mutex.RUnlock()
	
	if !connected || conn == nil {
		return fmt.Errorf("not connected to server")
	}
	
	// Build complete packet: [protocol][size][opcode][payload]
	packetSize := uint32(len(payload) + 1) // +1 for opcode
	packet := make([]byte, 0, 6+len(payload))
	
	packet = append(packet, protocol)
	
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, packetSize)
	packet = append(packet, sizeBytes...)
	
	packet = append(packet, opcode)
	packet = append(packet, payload...)
	
	if sc.debugLogging {
		fmt.Printf("[SERVER] → Sending packet to %s: protocol=0x%02x size=%d opcode=0x%02x\n", 
			sc.identifier, protocol, packetSize, opcode)
	}
	
	_, err := conn.Write(packet)
	return err
}

// handleServerStatus handles server status messages
func (sc *ServerConnection) handleServerStatus(payload []byte) {
	if sc.debugLogging {
		fmt.Printf("[SERVER] ← Server status from %s: %d bytes\n", sc.identifier, len(payload))
	}
}

// handleServerMessage handles server messages
func (sc *ServerConnection) handleServerMessage(payload []byte) {
	if len(payload) > 0 {
		message := string(payload)
		if sc.debugLogging {
			fmt.Printf("[SERVER] ← Server message from %s: %s\n", sc.identifier, message)
		}
	}
}

// handleIDChange handles client ID change notifications
func (sc *ServerConnection) handleIDChange(payload []byte) {
	if len(payload) >= 4 {
		newId := binary.LittleEndian.Uint32(payload[:4])
		sc.clientId = newId
		if sc.debugLogging {
			fmt.Printf("[SERVER] ← Client ID changed to %d (0x%08x) by %s\n", 
				newId, newId, sc.identifier)
		}
	}
}

// handleServerInfo handles server information
func (sc *ServerConnection) handleServerInfo(payload []byte) {
	if sc.debugLogging {
		fmt.Printf("[SERVER] ← Server info from %s: %d bytes\n", sc.identifier, len(payload))
	}
}

// IsConnected returns whether the server connection is active
func (sc *ServerConnection) IsConnected() bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return sc.connected && sc.handshakeDone
}

// Close closes the server connection
func (sc *ServerConnection) Close() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	
	if sc.connected {
		sc.connected = false
		sc.handshakeDone = false
		
		if sc.conn != nil {
			sc.conn.Close()
			sc.conn = nil
		}
		
		if sc.debugLogging {
			fmt.Printf("[SERVER] Connection closed to %s\n", sc.identifier)
		}
	}
}

// Ping sends a ping to keep connection alive
func (sc *ServerConnection) Ping() error {
	now := time.Now().Unix()
	if now - sc.lastPingTime < 120 { // Ping every 2 minutes
		return nil
	}
	
	sc.lastPingTime = now
	
	if sc.debugLogging {
		fmt.Printf("[SERVER] → Sending ping to %s\n", sc.identifier)
	}
	
	return sc.sendPacket(0xE3, 0x60, []byte{}) // Ping opcode
}