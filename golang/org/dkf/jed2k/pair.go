package jed2k

import (
	"fmt"
)

// Pair represents a C++ like pair template
type Pair[L, R any] struct {
	Left  L
	Right R
}

// NewPair creates a new pair
func NewPair[L, R any](left L, right R) *Pair[L, R] {
	return &Pair[L, R]{
		Left:  left,
		Right: right,
	}
}

// Make creates a new pair (alias for NewPair)
func Make[L, R any](left L, right R) *Pair[L, R] {
	return NewPair(left, right)
}

// GetLeft returns the left value
func (p *Pair[L, R]) GetLeft() L {
	return p.Left
}

// GetRight returns the right value
func (p *Pair[L, R]) GetRight() R {
	return p.Right
}

// Equals checks if two pairs are equal
func (p *Pair[L, R]) Equals(other *Pair[L, R]) bool {
	if other == nil {
		return false
	}
	// Note: This is a simplified comparison - for full compatibility with Java,
	// we would need to implement reflection-based equals checking
	return fmt.Sprintf("%v", p.Left) == fmt.Sprintf("%v", other.Left) &&
		fmt.Sprintf("%v", p.Right) == fmt.Sprintf("%v", other.Right)
}

// String returns string representation
func (p *Pair[L, R]) String() string {
	return fmt.Sprintf("Pair(left=%v, right=%v)", p.Left, p.Right)
}