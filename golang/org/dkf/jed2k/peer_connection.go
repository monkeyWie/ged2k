package jed2k

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol/client"
)

// ED2K packet opcodes
const (
	OP_HELLO          = 0x01
	OP_HELLOANSWER    = 0x4C
	OP_FILEREQUEST    = 0x58
	OP_FILEANSWER     = 0x59
	OP_REQUESTPARTS64 = 0x47
	OP_SENDINGPART64  = 0x46
)

// PendingBlock represents a block being downloaded
type PendingBlock struct {
	Block    *PieceBlock
	DataSize int64
	Data     []byte
}

// PeerConnection represents a connection to a peer using ed2k protocol
type PeerConnection struct {
	endpoint      *protocol.Endpoint
	peer          *Peer
	session       *Session
	transfer      *Transfer
	
	// Network connection
	conn          net.Conn
	connected     bool
	disconnecting bool
	failed        bool
	handshakeCompleted bool
	
	// Download state
	downloadQueue []*PendingBlock
	remotePieces  *protocol.BitField
	transferring  bool
	
	// Timing
	createTime    int64
	lastReceive   int64
	lastSend      int64
	
	// Statistics
	statistics    *Statistics
	
	// Synchronization
	mutex         sync.RWMutex
}

// NewPeerConnection creates a new peer connection
func NewPeerConnection(endpoint *protocol.Endpoint, session *Session) *PeerConnection {
	now := time.Now().Unix()
	return &PeerConnection{
		endpoint:      endpoint,
		session:       session,
		connected:     false,
		createTime:    now,
		lastReceive:   now,
		lastSend:      now,
		statistics:    NewStatistics(),
		downloadQueue: make([]*PendingBlock, 0),
		remotePieces:  protocol.NewBitFieldWithSize(0),
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

// SetTransfer sets the associated transfer
func (pc *PeerConnection) SetTransfer(transfer *Transfer) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pc.transfer = transfer
}

// GetTransfer returns the associated transfer
func (pc *PeerConnection) GetTransfer() *Transfer {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.transfer
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

// Connect establishes the connection using ed2k protocol
func (pc *PeerConnection) Connect() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.connected || pc.disconnecting {
		return nil
	}
	
	// Establish TCP connection
	addr := fmt.Sprintf("%s:%d", pc.endpoint.IPNet(), pc.endpoint.Port())
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		pc.failed = true
		return fmt.Errorf("failed to connect to %s: %v", addr, err)
	}
	
	pc.conn = conn
	pc.connected = true
	pc.lastReceive = time.Now().Unix()
	
	// Start the ed2k handshake
	go pc.handleConnection()
	
	return nil
}

// handleConnection handles the ed2k protocol communication with detailed logging
func (pc *PeerConnection) handleConnection() {
	defer pc.conn.Close()
	
	fmt.Printf("[PEER] → Starting ed2k handshake with %s\n", pc.endpoint.String())
	
	// Send Hello packet to initiate handshake
	if err := pc.sendHello(); err != nil {
		fmt.Printf("[PEER] ✗ Failed to send Hello to %s: %v\n", pc.endpoint.String(), err)
		pc.Close(exception.InternalError)
		return
	}
	
	fmt.Printf("[PEER] → Hello packet sent to %s\n", pc.endpoint.String())
	
	// Start reading packets
	buffer := make([]byte, 8192)
	packetBuffer := make([]byte, 0, 8192) // Buffer for accumulating packet data
	
	for pc.connected {
		pc.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := pc.conn.Read(buffer)
		if err != nil {
			fmt.Printf("[PEER] ✗ Read error from %s: %v\n", pc.endpoint.String(), err)
			pc.Close(exception.IOError)
			return
		}
		
		if n > 0 {
			pc.lastReceive = time.Now().Unix()
			packetBuffer = append(packetBuffer, buffer[:n]...)
			
			// Process complete packets
			for {
				processedBytes, err := pc.processReceivedDataWithLogging(packetBuffer)
				if err != nil {
					fmt.Printf("[PEER] ✗ Failed to process data from %s: %v\n", pc.endpoint.String(), err)
					pc.Close(exception.InternalError)
					return
				}
				
				if processedBytes == 0 {
					break // Need more data for a complete packet
				}
				
				// Remove processed bytes
				packetBuffer = packetBuffer[processedBytes:]
			}
		}
	}
}

// processReceivedDataWithLogging processes received packet data with protocol parsing and logging
func (pc *PeerConnection) processReceivedDataWithLogging(data []byte) (int, error) {
	if len(data) < 6 { // Minimum packet size: protocol(1) + size(4) + opcode(1)
		return 0, nil // Need more data
	}
	
	// Parse ED2K packet header
	protocolByte := data[0]
	packetSize := binary.LittleEndian.Uint32(data[1:5])
	opcode := data[5]
	
	if protocolByte != 0xE3 { // Standard ED2K protocol
		return len(data), fmt.Errorf("unsupported protocol: 0x%02x", protocolByte)
	}
	
	totalSize := int(packetSize) + 5 // +5 for protocol and size fields
	if len(data) < totalSize {
		return 0, nil // Need more data for complete packet
	}
	
	payload := data[6:totalSize]
	
	fmt.Printf("[PEER] ← Received packet from %s: opcode=0x%02x size=%d\n", 
		pc.endpoint.String(), opcode, packetSize)
	
	switch opcode {
	case OP_HELLOANSWER: // 0x4C
		err := pc.handleHelloAnswer(payload)
		if err != nil {
			return totalSize, fmt.Errorf("failed to handle HelloAnswer: %v", err)
		}
		
	case OP_FILEANSWER: // 0x59
		err := pc.handleFileAnswer(payload)
		if err != nil {
			return totalSize, fmt.Errorf("failed to handle FileAnswer: %v", err)
		}
		
	case OP_SENDINGPART64: // 0x46
		err := pc.handleSendingPart64(payload)
		if err != nil {
			return totalSize, fmt.Errorf("failed to handle SendingPart64: %v", err)
		}
		
	default:
		fmt.Printf("[PEER] ⚠ Unknown opcode 0x%02x from %s, ignoring\n", opcode, pc.endpoint.String())
	}
	
	return totalSize, nil
}

// handleHelloAnswer processes HelloAnswer packet completing handshake
func (pc *PeerConnection) handleHelloAnswer(payload []byte) error {
	if len(payload) < 23 { // Minimum: hash(16) + ID(4) + port(2) + tagcount(1)
		return fmt.Errorf("HelloAnswer payload too short: %d bytes", len(payload))
	}
	
	// Extract peer info from HelloAnswer
	peerHash := payload[0:16]
	clientID := binary.LittleEndian.Uint32(payload[16:20])
	port := binary.LittleEndian.Uint16(payload[20:22])
	
	fmt.Printf("[PEER] ← HelloAnswer from %s: ID=0x%08x port=%d hash=%x\n", 
		pc.endpoint.String(), clientID, port, peerHash[:4])
	
	pc.mutex.Lock()
	pc.handshakeCompleted = true
	pc.mutex.Unlock()
	
	// Now send file request
	return pc.sendFileRequestNew()
}

// sendFileRequestNew sends a file request for the transfer
func (pc *PeerConnection) sendFileRequestNew() error {
	if pc.transfer == nil {
		return fmt.Errorf("no transfer associated")
	}
	
	fmt.Printf("[PEER] → Sending FileRequest for %s to %s\n", 
		pc.transfer.name, pc.endpoint.String())
	
	fileReq := client.NewFileRequest(pc.transfer.hash)
	return pc.writePacket(OP_FILEREQUEST, fileReq)
}

// handleFileAnswer processes FileAnswer packet
func (pc *PeerConnection) handleFileAnswer(payload []byte) error {
	if len(payload) < 16 {
		return fmt.Errorf("FileAnswer payload too short")
	}
	
	// File hash
	fileHash := payload[0:16]
	
	fmt.Printf("[PEER] ← FileAnswer from %s for file hash %x\n", 
		pc.endpoint.String(), fileHash[:8])
	
	// Start requesting file parts
	return pc.requestBlocks()
}

// handleSendingPart64 processes received file data
func (pc *PeerConnection) handleSendingPart64(payload []byte) error {
	if len(payload) < 28 { // Minimum: hash(16) + start(8) + end(8)
		return fmt.Errorf("SendingPart64 payload too short")
	}
	
	fileHash := payload[0:16]
	startPos := binary.LittleEndian.Uint64(payload[16:24])
	endPos := binary.LittleEndian.Uint64(payload[24:32])
	actualData := payload[32:]
	
	dataSize := int64(len(actualData))
	expectedSize := int64(endPos - startPos + 1)
	
	fmt.Printf("[PEER] ← SendingPart64 from %s: hash=%x start=%d end=%d data=%d bytes\n", 
		pc.endpoint.String(), fileHash[:8], startPos, endPos, dataSize)
	
	if dataSize != expectedSize {
		fmt.Printf("[PEER] ⚠ Data size mismatch: expected %d, got %d\n", expectedSize, dataSize)
	}
	
	// Create block info for the received data
	if pc.transfer != nil {
		block := &PieceBlock{
			PieceIndex: int(startPos / 256000), // Assume 256KB pieces
			BlockIndex: int((startPos % 256000) / 176000), // Assume 176KB blocks  
		}
		
		// Convert block to actual data for the transfer
		blockData := actualData
		if err := pc.handleReceivedBlock(block, blockData); err != nil {
			return fmt.Errorf("failed to handle received block: %v", err)
		}
		
		fmt.Printf("[PEER] ✓ Block processed successfully from %s\n", pc.endpoint.String())
	}
	
	return nil
}

// sendHello sends the initial Hello packet
func (pc *PeerConnection) sendHello() error {
	// Create a user hash (in real implementation, this would be the session's user hash)
	userHashData := md5.Sum([]byte(fmt.Sprintf("user-%d", time.Now().Unix())))
	userHash, err := protocol.HashFromBytes(userHashData[:])
	if err != nil {
		return fmt.Errorf("failed to create user hash: %v", err)
	}
	
	// Create Hello packet
	hello := client.NewHello(userHash, 0x01234567, 4662) // Example client ID and port
	
	fmt.Printf("[PEER] → Sending Hello to %s: hash=%x ID=0x01234567 port=4662\n", 
		pc.endpoint.String(), userHash.Bytes()[:4])
	
	return pc.writePacket(OP_HELLO, hello)
}

// writePacket writes a packet to the connection
func (pc *PeerConnection) writePacket(opcode byte, packet protocol.Serializable) error {
	// ED2K packet format: [protocol(1)][size(4)][opcode(1)][data...]
	
	// Serialize packet data to buffer
	packetBuffer := &bytes.Buffer{}
	if err := packet.Put(packetBuffer); err != nil {
		return fmt.Errorf("failed to serialize packet: %v", err)
	}
	
	// Calculate total packet size
	dataSize := packetBuffer.Len()
	totalSize := 1 + 4 + 1 + dataSize // protocol + size + opcode + data
	
	buffer := make([]byte, 0, totalSize)
	finalBuffer := bytes.NewBuffer(buffer)
	
	// Write protocol version (0xE3 for standard ed2k)
	finalBuffer.WriteByte(0xE3)
	
	// Write packet size (size of opcode + data)
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(1+dataSize))
	finalBuffer.Write(sizeBytes)
	
	// Write opcode
	finalBuffer.WriteByte(opcode)
	
	// Write packet data
	finalBuffer.Write(packetBuffer.Bytes())
	
	// Send the packet
	_, err := pc.conn.Write(finalBuffer.Bytes())
	if err == nil {
		pc.lastSend = time.Now().Unix()
		fmt.Printf("[PEER] → Packet sent to %s: opcode=0x%02x size=%d\n", 
			pc.endpoint.String(), opcode, 1+dataSize)
	} else {
		fmt.Printf("[PEER] ✗ Failed to send packet to %s: %v\n", pc.endpoint.String(), err)
	}
	
	return err
}

// processReceivedData processes received packet data
func (pc *PeerConnection) processReceivedData(data []byte) error {
	// In a full implementation, this would:
	// 1. Parse the ED2K packet format
	// 2. Extract opcode and payload
	// 3. Handle different packet types (HelloAnswer, FileAnswer, SendingPart64, etc.)
	// 4. Update transfer state based on received data
	
	// For now, simulate receiving data for active transfers
	if pc.transfer != nil && len(pc.downloadQueue) > 0 {
		// Simulate receiving a block of data
		pendingBlock := pc.downloadQueue[0]
		pc.downloadQueue = pc.downloadQueue[1:]
		
		// Generate fake data for the block (in real implementation, this would be from the packet)
		blockData := make([]byte, pendingBlock.DataSize)
		for i := range blockData {
			blockData[i] = byte(i % 256)
		}
		
		// Process the received block
		if err := pc.handleReceivedBlock(pendingBlock.Block, blockData); err != nil {
			return fmt.Errorf("failed to handle received block: %v", err)
		}
	}
	
	return nil
}

// handleReceivedBlock processes a received file block
func (pc *PeerConnection) handleReceivedBlock(block *PieceBlock, data []byte) error {
	if pc.transfer == nil {
		return fmt.Errorf("no transfer associated with connection")
	}
	
	// Write block data to transfer
	if err := pc.transfer.WriteBlock(block, data); err != nil {
		return fmt.Errorf("failed to write block: %v", err)
	}
	
	// Update statistics
	pc.statistics.ReceiveBytes(0, int64(len(data))) // payload bytes
	
	return nil
}

// requestBlocks requests blocks from the peer
func (pc *PeerConnection) requestBlocks() error {
	if pc.transfer == nil || !pc.transfer.hasPicker() || len(pc.downloadQueue) > 0 {
		return nil
	}
	
	// Get blocks to download from the piece picker
	blocks := make([]*PieceBlock, 0, 3) // Request up to 3 blocks at a time
	pc.transfer.picker.PickPieces(&blocks, 3, pc.peer, 0)
	
	if len(blocks) == 0 {
		return nil // No blocks to request
	}
	
	// Create RequestParts64 packet
	reqParts := client.NewRequestParts64(pc.transfer.hash)
	
	for _, block := range blocks {
		startOffset, endOffset := block.Range(pc.transfer.size)
		reqParts.AddRequest(startOffset, endOffset)
		
		// Add to download queue
		pendingBlock := &PendingBlock{
			Block:    block,
			DataSize: int64(endOffset - startOffset),
		}
		pc.downloadQueue = append(pc.downloadQueue, pendingBlock)
	}
	
	// Send the request
	return pc.writePacket(OP_REQUESTPARTS64, reqParts)
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
	
	// Close the network connection
	if pc.conn != nil {
		pc.conn.Close()
		pc.conn = nil
	}
	
	pc.disconnecting = false
	
	return nil
}

// SecondTick is called every second to update the connection
func (pc *PeerConnection) SecondTick(tickIntervalMS int64) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.disconnecting || !pc.connected {
		return
	}
	
	currentTime := time.Now().Unix()
	
	// Check for timeouts
	if currentTime-pc.lastReceive > 30 { // 30 second timeout
		pc.Close(exception.ConnectionTimeout)
		return
	}
	
	// Request blocks if we need more data
	if pc.handshakeCompleted && !pc.transferring {
		go func() {
			if err := pc.requestBlocks(); err != nil {
				fmt.Printf("Failed to request blocks: %v\n", err)
			}
		}()
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