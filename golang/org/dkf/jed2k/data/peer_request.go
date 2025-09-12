package data

import (
	"fmt"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

// PeerRequest represents a request for data from a peer
type PeerRequest struct {
	Piece  int32 // piece index
	Start  int64 // start offset within piece
	Length int64 // length of data requested
}

// NewPeerRequest creates a new peer request
func NewPeerRequest(piece int32, start, length int64) *PeerRequest {
	return &PeerRequest{
		Piece:  piece,
		Start:  start,
		Length: length,
	}
}

// MakeRequest creates a peer request from begin and end positions
func MakeRequest(begin, end int64) (*PeerRequest, error) {
	if end <= begin || begin < 0 {
		return nil, fmt.Errorf("invalid peer request parameters: begin=%d, end=%d", begin, end)
	}
	
	piece := int32(begin / jed2k.PieceSize)
	start := begin % jed2k.PieceSize
	length := end - begin
	
	if length > jed2k.PieceSize {
		return nil, fmt.Errorf("peer request overflow: length=%d > piece_size=%d", length, jed2k.PieceSize)
	}
	
	return &PeerRequest{
		Piece:  piece,
		Start:  start,
		Length: length,
	}, nil
}

// MakeRequests creates multiple peer requests for a range
func MakeRequests(begin, end, fsize int64) ([]*PeerRequest, error) {
	// Clamp to file size
	if begin > fsize {
		begin = fsize
	}
	if end > fsize {
		end = fsize
	}
	
	var reqs []*PeerRequest
	
	for i := begin; i < end; {
		nextPieceBoundary := i + jed2k.PieceSize - (i % jed2k.PieceSize)
		requestEnd := end
		if nextPieceBoundary < end {
			requestEnd = nextPieceBoundary
		}
		
		pr, err := MakeRequest(i, requestEnd)
		if err != nil {
			return nil, err
		}
		
		reqs = append(reqs, pr)
		i += pr.Length
	}
	
	return reqs, nil
}

// MakeRequestFromPieceBlock creates a peer request from a piece block
func MakeRequestFromPieceBlock(b *PieceBlock, fsize int64) (*PeerRequest, error) {
	r := b.Range(fsize)
	return MakeRequest(r.Left, r.Right)
}

// InBlockOffset returns the offset within the block
func (pr *PeerRequest) InBlockOffset() int64 {
	return pr.Start % jed2k.BlockSize
}

// Range returns the file range this request covers
func (pr *PeerRequest) Range() *Range {
	begin := int64(pr.Piece)*jed2k.PieceSize + pr.Start
	end := begin + pr.Length
	return &Range{Left: begin, Right: end}
}

// Split splits the request into two parts at block boundary
func (pr *PeerRequest) Split() (*PeerRequest, *PeerRequest) {
	// Create first part (this request)
	first := &PeerRequest{
		Piece:  pr.Piece,
		Start:  pr.Start,
		Length: jed2k.BlockSize,
	}
	
	if first.Length > pr.Length {
		first.Length = pr.Length
	}
	
	// Create second part (remainder)
	second := &PeerRequest{
		Piece:  pr.Piece,
		Start:  pr.Start + first.Length,
		Length: pr.Length - first.Length,
	}
	
	return first, second
}

// String returns string representation
func (pr *PeerRequest) String() string {
	return fmt.Sprintf("piece %d start %d length %d", pr.Piece, pr.Start, pr.Length)
}