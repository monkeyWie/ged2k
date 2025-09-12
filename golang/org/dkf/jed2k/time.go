package jed2k

import (
	"sync/atomic"
	"time"
)

// Time provides cached time and high resolution time utilities
type Time struct {
	// currentCachedTime is updated every second or more frequently
	currentCachedTime int64
}

var timeInstance Time

// CurrentTime returns milliseconds from machine start, updated every second or more frequently
func CurrentTime() int64 {
	return atomic.LoadInt64(&timeInstance.currentCachedTime)
}

// UpdateCachedTime updates the global cached time
func UpdateCachedTime() {
	atomic.StoreInt64(&timeInstance.currentCachedTime, CurrentTimeHiRes())
}

// CurrentTimeHiRes returns actual time on every call in milliseconds
func CurrentTimeHiRes() int64 {
	return time.Now().UnixNano() / 1000000
}

// CurrentTimeMillis returns current time offset convertible to Date object
func CurrentTimeMillis() int64 {
	return time.Now().UnixMilli()
}

// Minutes converts minutes to milliseconds
func Minutes(value int) int64 {
	return int64(value * 1000 * 60)
}

// Hours converts hours to milliseconds
func Hours(value int) int64 {
	return int64(value * 1000 * 3600)
}

// Seconds converts seconds to milliseconds
func Seconds(value int) int64 {
	return int64(value * 1000)
}

func init() {
	// Initialize the cached time
	UpdateCachedTime()
}