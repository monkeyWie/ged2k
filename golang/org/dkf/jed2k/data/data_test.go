package data

import (
	"testing"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

func TestRangeBasic(t *testing.T) {
	r := NewRange(10, 20)
	
	if r.Left != 10 || r.Right != 20 {
		t.Errorf("Expected range [10..20], got [%d..%d]", r.Left, r.Right)
	}
	
	if r.Size() != 10 {
		t.Errorf("Expected size 10, got %d", r.Size())
	}
	
	if !r.Contains(15) {
		t.Error("Range should contain 15")
	}
	
	if r.Contains(25) {
		t.Error("Range should not contain 25")
	}
}

func TestRangeCompareTo(t *testing.T) {
	r1 := NewRange(10, 20)
	r2 := NewRange(10, 20)
	r3 := NewRange(15, 25)
	
	if r1.CompareTo(r2) != 0 {
		t.Error("Equal ranges should compare to 0")
	}
	
	if r1.CompareTo(r3) >= 0 {
		t.Error("r1 should be less than r3")
	}
}

func TestPeerRequestMake(t *testing.T) {
	// Test basic request creation
	pr, err := MakeRequest(0, jed2k.PieceSize)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	if pr.Piece != 0 {
		t.Errorf("Expected piece 0, got %d", pr.Piece)
	}
	
	if pr.Start != 0 {
		t.Errorf("Expected start 0, got %d", pr.Start)
	}
	
	if pr.Length != jed2k.PieceSize {
		t.Errorf("Expected length %d, got %d", jed2k.PieceSize, pr.Length)
	}
}

func TestPeerRequestInvalid(t *testing.T) {
	// Test invalid parameters
	_, err := MakeRequest(100, 50)
	if err == nil {
		t.Error("Expected error for invalid begin/end")
	}
	
	_, err = MakeRequest(-1, 100)
	if err == nil {
		t.Error("Expected error for negative begin")
	}
}

func TestPeerRequestMultiple(t *testing.T) {
	// Test multiple requests
	reqs, err := MakeRequests(0, jed2k.PieceSize*2, jed2k.PieceSize*3)
	if err != nil {
		t.Fatalf("Failed to create requests: %v", err)
	}
	
	if len(reqs) != 2 {
		t.Errorf("Expected 2 requests, got %d", len(reqs))
	}
	
	// First request should be for piece 0
	if reqs[0].Piece != 0 {
		t.Errorf("Expected first request for piece 0, got %d", reqs[0].Piece)
	}
	
	// Second request should be for piece 1
	if reqs[1].Piece != 1 {
		t.Errorf("Expected second request for piece 1, got %d", reqs[1].Piece)
	}
}

func TestPieceBlockBasic(t *testing.T) {
	pb := NewPieceBlockWithValues(1, 2)
	
	if pb.PieceIndex != 1 {
		t.Errorf("Expected piece index 1, got %d", pb.PieceIndex)
	}
	
	if pb.PieceBlock != 2 {
		t.Errorf("Expected block index 2, got %d", pb.PieceBlock)
	}
	
	expected := int64(1)*int64(jed2k.BlocksPerPiece) + 2
	if pb.BlocksOffset() != expected {
		t.Errorf("Expected blocks offset %d, got %d", expected, pb.BlocksOffset())
	}
}

func TestPieceBlockMake(t *testing.T) {
	// Test creating block from offset
	offset := jed2k.PieceSize + jed2k.BlockSize*2
	pb := MakeBlock(offset)
	
	if pb.PieceIndex != 1 {
		t.Errorf("Expected piece index 1, got %d", pb.PieceIndex)
	}
	
	if pb.PieceBlock != 2 {
		t.Errorf("Expected block index 2, got %d", pb.PieceBlock)
	}
}

func TestPieceBlockRange(t *testing.T) {
	pb := NewPieceBlockWithValues(0, 0)
	r := pb.Range(jed2k.PieceSize)
	
	if r.Left != 0 {
		t.Errorf("Expected range start 0, got %d", r.Left)
	}
	
	if r.Right != jed2k.BlockSize {
		t.Errorf("Expected range end %d, got %d", jed2k.BlockSize, r.Right)
	}
}