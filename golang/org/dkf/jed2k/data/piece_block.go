package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

// PieceBlock describes a block of data in a piece
// Each piece contains few blocks, from 1 to BLOCKS_PER_PIECE constant
type PieceBlock struct {
	PieceIndex int32 // piece index
	PieceBlock int32 // block index within piece
}

// NewPieceBlock creates a new piece block
func NewPieceBlock() *PieceBlock {
	return &PieceBlock{
		PieceIndex: -1,
		PieceBlock: -1,
	}
}

// NewPieceBlockWithValues creates a new piece block with specific values
func NewPieceBlockWithValues(pieceIndex, pieceBlock int32) *PieceBlock {
	if pieceIndex < 0 || pieceBlock < 0 {
		panic("piece index and block index must be non-negative")
	}
	return &PieceBlock{
		PieceIndex: pieceIndex,
		PieceBlock: pieceBlock,
	}
}

// BlocksOffset returns the offset of current block in blocks
func (pb *PieceBlock) BlocksOffset() int64 {
	return int64(pb.PieceIndex)*int64(jed2k.BlocksPerPiece) + int64(pb.PieceBlock)
}

// MakeBlockFromRequest creates a piece block from a peer request
func MakeBlockFromRequest(r *PeerRequest) *PieceBlock {
	return NewPieceBlockWithValues(r.Piece, int32(r.Start/jed2k.BlockSize))
}

// MakeBlock creates a piece block from file offset
func MakeBlock(offset int64) *PieceBlock {
	piece := int32(offset / jed2k.PieceSize)
	start := offset % jed2k.PieceSize
	return NewPieceBlockWithValues(piece, int32(start/jed2k.BlockSize))
}

// Range returns the range which this block covers
func (pb *PieceBlock) Range(size int64) *Range {
	begin := int64(pb.PieceIndex)*jed2k.PieceSize + int64(pb.PieceBlock)*jed2k.BlockSize
	normalEnd := int64(pb.PieceIndex)*jed2k.PieceSize + int64(pb.PieceBlock+1)*jed2k.BlockSize
	
	end := begin + jed2k.BlockSize
	if normalEnd < end {
		end = normalEnd
	}
	if size < end {
		end = size
	}
	
	return &Range{Left: begin, Right: end}
}

// Size returns the block size in bytes
func (pb *PieceBlock) Size(totalSize int64) int32 {
	r := pb.Range(totalSize)
	return int32(r.Right - r.Left)
}

// CompareTo compares with another piece block
func (pb *PieceBlock) CompareTo(other *PieceBlock) int {
	offset1 := pb.BlocksOffset()
	offset2 := other.BlocksOffset()
	
	if offset1 < offset2 {
		return -1
	} else if offset1 > offset2 {
		return 1
	}
	return 0
}

// Equals checks equality with another piece block
func (pb *PieceBlock) Equals(other *PieceBlock) bool {
	return pb.CompareTo(other) == 0
}

// HashCode returns hash code for the piece block
func (pb *PieceBlock) HashCode() int32 {
	return pb.PieceIndex*int32(jed2k.BlocksPerPiece) + pb.PieceBlock
}

// Get deserializes from buffer
func (pb *PieceBlock) Get(src *bytes.Buffer) error {
	if src.Len() < 8 {
		return fmt.Errorf("buffer underflow: need 8 bytes, have %d", src.Len())
	}
	
	err := binary.Read(src, binary.LittleEndian, &pb.PieceIndex)
	if err != nil {
		return err
	}
	
	return binary.Read(src, binary.LittleEndian, &pb.PieceBlock)
}

// Put serializes to buffer
func (pb *PieceBlock) Put(dst *bytes.Buffer) error {
	err := binary.Write(dst, binary.LittleEndian, pb.PieceIndex)
	if err != nil {
		return err
	}
	
	return binary.Write(dst, binary.LittleEndian, pb.PieceBlock)
}

// BytesCount returns the number of bytes this type uses
func (pb *PieceBlock) BytesCount() int {
	return 8 // 4 bytes for piece index + 4 bytes for block index
}

// String returns string representation
func (pb *PieceBlock) String() string {
	return fmt.Sprintf("piece{%d} block{%d}", pb.PieceIndex, pb.PieceBlock)
}