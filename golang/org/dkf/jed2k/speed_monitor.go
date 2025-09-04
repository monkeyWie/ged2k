package jed2k

const (
	INVALID_SPEED = -1
	INVALID_ETA   = -1
)

// SpeedMonitor monitors and calculates average speed
type SpeedMonitor struct {
	speedSamples []int64
	roundRobin   int
	totalSamples int
}

// NewSpeedMonitor creates a new speed monitor with given sample limit
func NewSpeedMonitor(samplesLimit int) *SpeedMonitor {
	return &SpeedMonitor{
		speedSamples: make([]int64, samplesLimit),
		roundRobin:   0,
		totalSamples: 0,
	}
}

// AddSample adds a speed sample to the monitor
func (sm *SpeedMonitor) AddSample(speedSample int64) {
	if sm.roundRobin == len(sm.speedSamples) {
		sm.roundRobin = 0
		if sm.totalSamples != len(sm.speedSamples) {
			panic("totalSamples should equal speedSamples length")
		}
	}

	if sm.roundRobin < len(sm.speedSamples) {
		sm.speedSamples[sm.roundRobin] = speedSample
		sm.roundRobin++
	}

	if sm.totalSamples != len(sm.speedSamples) {
		sm.totalSamples++
	}
}

// AverageSpeed calculates the average speed from all samples
func (sm *SpeedMonitor) AverageSpeed() int64 {
	if sm.totalSamples == 0 {
		return INVALID_SPEED
	}

	sum := int64(0)
	for i := 0; i < sm.totalSamples; i++ {
		sum += sm.speedSamples[i]
	}

	return sum / int64(sm.totalSamples)
}

// GetNumSamples returns the number of samples collected
func (sm *SpeedMonitor) GetNumSamples() int {
	return sm.totalSamples
}

// Clear resets the speed monitor
func (sm *SpeedMonitor) Clear() {
	sm.roundRobin = 0
	sm.totalSamples = 0
}