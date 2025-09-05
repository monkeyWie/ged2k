package jed2k

import (
	"sync"
	
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// PieceBlock represents a block within a piece
type PieceBlock struct {
	PieceIndex int
	BlockIndex int
}

// NewPieceBlock creates a new piece block
func NewPieceBlock(pieceIndex, blockIndex int) *PieceBlock {
	return &PieceBlock{
		PieceIndex: pieceIndex,
		BlockIndex: blockIndex,
	}
}

// Range returns the byte range for this block in the file
func (pb *PieceBlock) Range(fileSize int64) (startOffset, endOffset uint64) {
	startOffset = uint64(pb.PieceIndex)*uint64(PieceSize) + uint64(pb.BlockIndex)*uint64(BlockSize)
	endOffset = startOffset + uint64(BlockSize)
	
	// Ensure we don't go past the file size
	if endOffset > uint64(fileSize) {
		endOffset = uint64(fileSize)
	}
	
	return startOffset, endOffset
}

// DownloadingPiece represents a piece being downloaded
type DownloadingPiece struct {
	PieceIndex   int
	BlockCount   int
	FinishedMask *protocol.BitField // tracks which blocks are complete
	mutex        sync.RWMutex
}

// NewDownloadingPiece creates a new downloading piece
func NewDownloadingPiece(pieceIndex, blockCount int) *DownloadingPiece {
	return &DownloadingPiece{
		PieceIndex:   pieceIndex,
		BlockCount:   blockCount,
		FinishedMask: protocol.NewBitFieldWithSize(blockCount),
	}
}

// IsFinished returns true if the specified block is finished
func (dp *DownloadingPiece) IsFinished(blockIndex int) bool {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()
	return dp.FinishedMask.GetBit(blockIndex)
}

// MarkFinished marks a block as finished
func (dp *DownloadingPiece) MarkFinished(blockIndex int) {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()
	dp.FinishedMask.SetBit(blockIndex)
}

// IsComplete returns true if all blocks in this piece are finished
func (dp *DownloadingPiece) IsComplete() bool {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()
	return dp.FinishedMask.Count() == dp.BlockCount
}

// PiecePicker manages piece/block selection for downloads
type PiecePicker struct {
	numPieces        int
	blocksInLastPiece int
	havePieces       *protocol.BitField
	downloadingPieces map[int]*DownloadingPiece
	mutex            sync.RWMutex
}

// NewPiecePicker creates a new piece picker
func NewPiecePicker(numPieces, blocksInLastPiece int) *PiecePicker {
	return &PiecePicker{
		numPieces:        numPieces,
		blocksInLastPiece: blocksInLastPiece,
		havePieces:       protocol.NewBitFieldWithSize(numPieces),
		downloadingPieces: make(map[int]*DownloadingPiece),
	}
}

// NumPieces returns the total number of pieces
func (pp *PiecePicker) NumPieces() int {
	return pp.numPieces
}

// NumHave returns the number of pieces we have
func (pp *PiecePicker) NumHave() int {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()
	return pp.havePieces.Count()
}

// HavePiece returns true if we have the specified piece
func (pp *PiecePicker) HavePiece(pieceIndex int) bool {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()
	return pp.havePieces.GetBit(pieceIndex)
}

// WeHave marks a piece as completed
func (pp *PiecePicker) WeHave(pieceIndex int) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	
	pp.havePieces.SetBit(pieceIndex)
	delete(pp.downloadingPieces, pieceIndex)
}

// RestoreHave marks a piece as completed (for resume data)
func (pp *PiecePicker) RestoreHave(pieceIndex int) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.havePieces.SetBit(pieceIndex)
}

// PickPieces selects blocks to download for a peer
func (pp *PiecePicker) PickPieces(blocks *[]*PieceBlock, maxBlocks int, peer *Peer, speed int) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	
	// Find pieces that need downloading
	for pieceIndex := 0; pieceIndex < pp.numPieces && len(*blocks) < maxBlocks; pieceIndex++ {
		// Skip pieces we already have
		if pp.havePieces.GetBit(pieceIndex) {
			continue
		}
		
		// Get or create downloading piece
		dp, exists := pp.downloadingPieces[pieceIndex]
		if !exists {
			blockCount := pp.getBlockCountForPiece(pieceIndex)
			dp = NewDownloadingPiece(pieceIndex, blockCount)
			pp.downloadingPieces[pieceIndex] = dp
		}
		
		// Find blocks in this piece that need downloading
		for blockIndex := 0; blockIndex < dp.BlockCount && len(*blocks) < maxBlocks; blockIndex++ {
			if !dp.IsFinished(blockIndex) {
				pb := NewPieceBlock(pieceIndex, blockIndex)
				*blocks = append(*blocks, pb)
			}
		}
	}
}

// DownloadPiece marks a piece as being downloaded
func (pp *PiecePicker) DownloadPiece(pieceIndex int) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	
	if _, exists := pp.downloadingPieces[pieceIndex]; !exists {
		blockCount := pp.getBlockCountForPiece(pieceIndex)
		pp.downloadingPieces[pieceIndex] = NewDownloadingPiece(pieceIndex, blockCount)
	}
}

// MarkAsFinished marks a block as finished
func (pp *PiecePicker) MarkAsFinished(pb *PieceBlock) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	
	if dp, exists := pp.downloadingPieces[pb.PieceIndex]; exists {
		dp.MarkFinished(pb.BlockIndex)
		
		// Check if piece is complete
		if dp.IsComplete() {
			pp.havePieces.SetBit(pb.PieceIndex)
			delete(pp.downloadingPieces, pb.PieceIndex)
		}
	}
}

// AbortDownload aborts download of a block
func (pp *PiecePicker) AbortDownload(pb *PieceBlock, peer *Peer) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	
	// For now, just remove the block from downloading state
	// In a full implementation, we'd track which peer is downloading each block
}

// GetDownloadingQueue returns the list of downloading pieces
func (pp *PiecePicker) GetDownloadingQueue() []*DownloadingPiece {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()
	
	queue := make([]*DownloadingPiece, 0, len(pp.downloadingPieces))
	for _, dp := range pp.downloadingPieces {
		queue = append(queue, dp)
	}
	return queue
}

// getBlockCountForPiece returns the number of blocks in a piece
func (pp *PiecePicker) getBlockCountForPiece(pieceIndex int) int {
	if pieceIndex == pp.numPieces-1 {
		return pp.blocksInLastPiece
	}
	return int(PieceSize / BlockSize)
}