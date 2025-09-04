package data

import "fmt"

// Range represents a range with left and right bounds
type Range struct {
	Left  int64
	Right int64
}

// NewRange creates a new range
func NewRange(left, right int64) *Range {
	if right <= left {
		panic("right must be greater than left")
	}
	return &Range{Left: left, Right: right}
}

// MakeRange creates a new range (factory method)
func MakeRange(left, right int64) *Range {
	return NewRange(left, right)
}

// String returns string representation
func (r *Range) String() string {
	return fmt.Sprintf("range [%d..%d]", r.Left, r.Right)
}

// Equals checks equality with another range
func (r *Range) Equals(other *Range) bool {
	return r.Left == other.Left && r.Right == other.Right
}

// CompareTo compares with another range
func (r *Range) CompareTo(other *Range) int {
	if r.Left != other.Left {
		if r.Left < other.Left {
			return -1
		}
		if r.Left > other.Left {
			return 1
		}
	}
	
	if r.Right < other.Right {
		return -1
	}
	if r.Right > other.Right {
		return 1
	}
	return 0
}

// Contains checks if a value is within the range
func (r *Range) Contains(value int64) bool {
	return value >= r.Left && value < r.Right
}

// Size returns the size of the range
func (r *Range) Size() int64 {
	return r.Right - r.Left
}

// IsEmpty checks if the range is empty
func (r *Range) IsEmpty() bool {
	return r.Left >= r.Right
}

// Overlaps checks if this range overlaps with another
func (r *Range) Overlaps(other *Range) bool {
	return r.Left < other.Right && other.Left < r.Right
}