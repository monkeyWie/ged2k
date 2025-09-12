package pool

import (
	"container/list"
	"fmt"
	"math"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
)

// Pool provides generic object pooling functionality
type Pool[T any] struct {
	maxBuffersCount      int
	allocatedBuffersCount int
	maxAllocatedCount    int
	freeBuffers         *list.List // List of T
	bufferReleaseTimes  *list.List // List of int64
	createFunc          func() (T, error)
}

// NewPool creates a new pool with maximum buffer count
func NewPool[T any](maxBuffers int, createFunc func() (T, error)) *Pool[T] {
	if maxBuffers <= 0 {
		panic("maxBuffers must be > 0")
	}
	
	return &Pool[T]{
		maxBuffersCount:     maxBuffers,
		freeBuffers:        list.New(),
		bufferReleaseTimes: list.New(),
		createFunc:         createFunc,
	}
}

// Allocate gets an object from the pool or creates a new one
func (p *Pool[T]) Allocate() (T, error) {
	var result T
	
	// Try to get from free buffers
	if p.freeBuffers.Len() > 0 {
		front := p.freeBuffers.Front()
		result = front.Value.(T)
		p.freeBuffers.Remove(front)
		
		// Remove corresponding release time
		if p.bufferReleaseTimes.Len() > 0 {
			p.bufferReleaseTimes.Remove(p.bufferReleaseTimes.Front())
		}
	}
	
	// Verify list consistency
	if p.freeBuffers.Len() != p.bufferReleaseTimes.Len() {
		panic("freeBuffers and bufferReleaseTimes lists must have same size")
	}
	
	// If no free buffer and we can allocate more
	if p.freeBuffers.Len() == 0 && p.allocatedBuffersCount < p.maxBuffersCount {
		var err error
		result, err = p.createFunc()
		if err != nil {
			return result, err
		}
	}
	
	// Check if we have a valid result
	if p.freeBuffers.Len() == 0 && p.allocatedBuffersCount >= p.maxBuffersCount {
		return result, exception.NewJED2KException(exception.NoMemory)
	}
	
	p.allocatedBuffersCount++
	p.maxAllocatedCount = int(math.Max(float64(p.maxAllocatedCount), float64(p.allocatedBuffersCount)))
	
	return result, nil
}

// Deallocate returns an object to the pool
func (p *Pool[T]) Deallocate(obj T, sessionTime int64) {
	if p.freeBuffers.Len() != p.bufferReleaseTimes.Len() {
		panic("freeBuffers and bufferReleaseTimes lists must have same size")
	}
	
	p.allocatedBuffersCount--
	
	// Add to cache only if limit not exceeded
	if p.maxBuffersCount > p.allocatedBuffersCount+p.GetCachedBuffersCount() {
		p.freeBuffers.PushFront(obj)
		p.bufferReleaseTimes.PushFront(sessionTime)
	}
}

// GetCachedBuffersCount returns number of cached free buffers
func (p *Pool[T]) GetCachedBuffersCount() int {
	return p.freeBuffers.Len()
}

// ReduceCache reduces the cache size to specified value
func (p *Pool[T]) ReduceCache(cacheSize int) int {
	if cacheSize == 0 {
		p.freeBuffers = list.New()
		p.bufferReleaseTimes = list.New()
	} else {
		for p.freeBuffers.Len() > cacheSize {
			p.freeBuffers.Remove(p.freeBuffers.Front())
			p.bufferReleaseTimes.Remove(p.bufferReleaseTimes.Front())
		}
	}
	
	return p.freeBuffers.Len()
}

// SecondTick performs periodic maintenance
func (p *Pool[T]) SecondTick(currentSessionTime int64) {
	// Free obsolete buffers - implementation can be added as needed
}

// SetMaxBuffersCount sets the maximum buffer count
func (p *Pool[T]) SetMaxBuffersCount(maxBuffers int) {
	if maxBuffers <= 0 {
		panic("maxBuffers must be > 0")
	}
	
	if maxBuffers < p.maxBuffersCount {
		// Reduce cache if new limit is lower
		newCacheSize := int(math.Max(float64(p.GetCachedBuffersCount()-(p.maxBuffersCount-maxBuffers)), 0))
		p.ReduceCache(newCacheSize)
	}
	
	p.maxBuffersCount = maxBuffers
}

// String returns string representation of the pool
func (p *Pool[T]) String() string {
	return fmt.Sprintf("buffer pool max{%d} allocated/maxallocated {%d/%d} free {%d}",
		p.maxBuffersCount, p.allocatedBuffersCount, p.maxAllocatedCount, p.freeBuffers.Len())
}

// GetMaxBuffersCount returns maximum buffer count
func (p *Pool[T]) GetMaxBuffersCount() int {
	return p.maxBuffersCount
}

// GetAllocatedBuffersCount returns currently allocated buffer count
func (p *Pool[T]) GetAllocatedBuffersCount() int {
	return p.allocatedBuffersCount
}

// GetMaxAllocatedCount returns maximum allocated count reached
func (p *Pool[T]) GetMaxAllocatedCount() int {
	return p.maxAllocatedCount
}