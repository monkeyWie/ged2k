package pool

import (
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/exception"
)

// BufferPool provides ByteBuffer pooling functionality
type BufferPool struct {
	*Pool[[]byte]
}

// NewBufferPool creates a new buffer pool
func NewBufferPool(maxBuffers int) *BufferPool {
	createFunc := func() ([]byte, error) {
		// Use BlockSizeInt from Constants
		blockSize := jed2k.BlockSizeInt
		
		buffer := make([]byte, blockSize)
		if buffer == nil {
			return nil, exception.NewJED2KException(exception.NoMemory)
		}
		
		return buffer, nil
	}
	
	return &BufferPool{
		Pool: NewPool(maxBuffers, createFunc),
	}
}

// Deallocate returns a buffer to the pool after clearing it
func (bp *BufferPool) Deallocate(buffer []byte, sessionTime int64) {
	if buffer == nil {
		panic("buffer cannot be nil")
	}
	
	// Clear the buffer (set all bytes to 0)
	for i := range buffer {
		buffer[i] = 0
	}
	
	bp.Pool.Deallocate(buffer, sessionTime)
}