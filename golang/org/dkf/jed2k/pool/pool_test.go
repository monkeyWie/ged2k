package pool

import (
	"testing"
)

func TestPool(t *testing.T) {
	// Test creating a pool of byte slices
	createFunc := func() ([]byte, error) {
		return make([]byte, 1024), nil
	}
	
	pool := NewPool(5, createFunc) // Max 5 buffers
	
	// Test initial state
	if pool.GetMaxBuffersCount() != 5 {
		t.Error("Max buffers count should be 5")
	}
	
	if pool.GetAllocatedBuffersCount() != 0 {
		t.Error("Initially allocated count should be 0")
	}
	
	if pool.GetCachedBuffersCount() != 0 {
		t.Error("Initially cached count should be 0")
	}
	
	// Test allocation
	buf1, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to allocate buffer: %v", err)
	}
	
	if len(buf1) != 1024 {
		t.Error("Buffer should be 1024 bytes")
	}
	
	if pool.GetAllocatedBuffersCount() != 1 {
		t.Error("Allocated count should be 1")
	}
	
	// Allocate more buffers
	buf2, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to allocate second buffer: %v", err)
	}
	
	if pool.GetAllocatedBuffersCount() != 2 {
		t.Error("Allocated count should be 2")
	}
	
	// Test deallocation
	pool.Deallocate(buf1, 1000)
	if pool.GetAllocatedBuffersCount() != 1 {
		t.Error("Allocated count should be 1 after deallocation")
	}
	
	if pool.GetCachedBuffersCount() != 1 {
		t.Error("Cached count should be 1 after deallocation")
	}
	
	// Reallocate should reuse cached buffer
	buf3, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to reallocate buffer: %v", err)
	}
	
	if pool.GetCachedBuffersCount() != 0 {
		t.Error("Cached count should be 0 after reallocation")
	}
	
	if pool.GetAllocatedBuffersCount() != 2 {
		t.Error("Allocated count should be 2 after reallocation")
	}
	
	// Clean up
	pool.Deallocate(buf2, 2000)
	pool.Deallocate(buf3, 3000)
}

func TestPoolLimits(t *testing.T) {
	createFunc := func() ([]byte, error) {
		return make([]byte, 100), nil
	}
	
	pool := NewPool(2, createFunc) // Max 2 buffers
	
	// Allocate max buffers
	buf1, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to allocate first buffer: %v", err)
	}
	
	buf2, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to allocate second buffer: %v", err)
	}
	
	// Third allocation should fail
	_, err = pool.Allocate()
	if err == nil {
		t.Error("Third allocation should fail when limit reached")
	}
	
	// After deallocation, should be able to allocate again
	pool.Deallocate(buf1, 1000)
	buf3, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Should be able to allocate after deallocation: %v", err)
	}
	
	// Clean up
	pool.Deallocate(buf2, 2000)
	pool.Deallocate(buf3, 3000)
}

func TestBufferPool(t *testing.T) {
	pool := NewBufferPool(3) // Max 3 buffers
	
	// Test allocation
	buf1, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to allocate buffer: %v", err)
	}
	
	if len(buf1) == 0 {
		t.Error("Buffer should have non-zero length")
	}
	
	// Modify buffer to test clearing
	buf1[0] = 255
	buf1[len(buf1)-1] = 128
	
	// Test deallocation with clearing
	pool.Deallocate(buf1, 1000)
	
	// Reallocate and check if cleared
	buf2, err := pool.Allocate()
	if err != nil {
		t.Fatalf("Failed to reallocate buffer: %v", err)
	}
	
	// Buffer should be cleared (all zeros)
	for i, b := range buf2 {
		if b != 0 {
			t.Errorf("Buffer should be cleared, but byte at %d is %d", i, b)
		}
	}
	
	// Clean up
	pool.Deallocate(buf2, 2000)
}

func TestPoolResize(t *testing.T) {
	createFunc := func() ([]byte, error) {
		return make([]byte, 100), nil
	}
	
	pool := NewPool(5, createFunc)
	
	// Allocate some buffers
	buf1, _ := pool.Allocate()
	buf2, _ := pool.Allocate()
	pool.Deallocate(buf1, 1000)
	pool.Deallocate(buf2, 2000)
	
	if pool.GetCachedBuffersCount() != 2 {
		t.Error("Should have 2 cached buffers")
	}
	
	// Reduce max buffer count
	pool.SetMaxBuffersCount(3)
	if pool.GetMaxBuffersCount() != 3 {
		t.Error("Max buffers count should be updated to 3")
	}
	
	// Cache should be reduced if necessary
	if pool.GetCachedBuffersCount() > 3 {
		t.Error("Cache should not exceed new max limit")
	}
}

func TestPoolSecondTick(t *testing.T) {
	createFunc := func() ([]byte, error) {
		return make([]byte, 100), nil
	}
	
	pool := NewPool(5, createFunc)
	
	// SecondTick should not panic
	pool.SecondTick(1000)
	
	// Test with some buffers
	buf, _ := pool.Allocate()
	pool.Deallocate(buf, 1000)
	pool.SecondTick(2000)
}

func TestPoolString(t *testing.T) {
	createFunc := func() ([]byte, error) {
		return make([]byte, 100), nil
	}
	
	pool := NewPool(5, createFunc)
	str := pool.String()
	
	if len(str) == 0 {
		t.Error("Pool string representation should not be empty")
	}
}