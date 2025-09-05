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

// handleConnection handles the ed2k protocol communication
func (pc *PeerConnection) handleConnection() {
	defer pc.conn.Close()
	
	// Send Hello packet to initiate handshake
	if err := pc.sendHello(); err != nil {
		fmt.Printf("Failed to send Hello: %v\n", err)
		pc.Close(exception.InternalError)
		return
	}
	
	// Start reading packets
	buffer := make([]byte, 8192)
	for pc.connected {
		pc.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := pc.conn.Read(buffer)
		if err != nil {
			pc.Close(exception.IOError)
			return
		}
		
		if n > 0 {
			pc.lastReceive = time.Now().Unix()
			if err := pc.processReceivedData(buffer[:n]); err != nil {
				fmt.Printf("Failed to process data: %v\n", err)
				pc.Close(exception.InternalError)
				return
			}
		}
	}
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