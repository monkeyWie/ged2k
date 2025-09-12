package jed2k

import "container/list"

// StatChannel represents a statistics channel
type StatChannel struct {
	secondCounter int64
	totalCounter  int64
	average5Sec   int64
	average30Sec  int64
	samples       *list.List // LinkedList of 5 int64 values
}

// NewStatChannel creates a new statistics channel
func NewStatChannel() *StatChannel {
	sc := &StatChannel{
		samples: list.New(),
	}
	// Initialize with 5 zero samples
	for i := 0; i < 5; i++ {
		sc.samples.PushBack(int64(0))
	}
	return sc
}

// Add adds bytes to the counter
func (sc *StatChannel) Add(count int64) {
	if count < 0 {
		panic("count must be >= 0")
	}
	sc.secondCounter += count
	sc.totalCounter += count
}

// Counter returns the current second counter
func (sc *StatChannel) Counter() int64 {
	return sc.secondCounter
}

// Total returns the total counter
func (sc *StatChannel) Total() int64 {
	return sc.totalCounter
}

// Rate returns the 5-second average rate
func (sc *StatChannel) Rate() int64 {
	return sc.average5Sec
}

// LowPassRate returns the 30-second low pass rate
func (sc *StatChannel) LowPassRate() int64 {
	return sc.average30Sec
}

// Clear resets all counters
func (sc *StatChannel) Clear() {
	sc.secondCounter = 0
	sc.totalCounter = 0
	sc.average5Sec = 0
	sc.average30Sec = 0
	sc.samples = list.New()
	// Re-initialize with 5 zero samples
	for i := 0; i < 5; i++ {
		sc.samples.PushBack(int64(0))
	}
}

// AddChannel adds another StatChannel to this one
func (sc *StatChannel) AddChannel(s *StatChannel) *StatChannel {
	if sc.secondCounter < 0 || sc.totalCounter < 0 || s.secondCounter < 0 {
		panic("counters must be >= 0")
	}
	sc.secondCounter += s.secondCounter
	sc.totalCounter += s.totalCounter
	if sc.secondCounter < 0 || sc.totalCounter < 0 {
		panic("counter overflow")
	}
	return sc
}

// SecondTick updates statistics on each second tick
func (sc *StatChannel) SecondTick(timeIntervalMS int64) {
	sample := (sc.secondCounter * 1000) / timeIntervalMS
	if sample < 0 {
		panic("sample must be >= 0")
	}
	
	// Push new sample and remove the oldest
	sc.samples.PushFront(sample)
	sc.samples.Remove(sc.samples.Back())
	
	// Calculate 5-second average
	sum := int64(0)
	count := 0
	for e := sc.samples.Front(); e != nil; e = e.Next() {
		sum += e.Value.(int64)
		count++
	}
	
	if count != 5 {
		panic("samples list must have exactly 5 elements")
	}
	
	sc.average5Sec = sum / 5
	sc.average30Sec = sc.average30Sec*29/30 + sample/30
	sc.secondCounter = 0
}

// ChannelType represents the type of statistics channel
type ChannelType int

const (
	UPLOAD_PAYLOAD ChannelType = iota
	UPLOAD_PROTOCOL
	DOWNLOAD_PAYLOAD
	DOWNLOAD_PROTOCOL
	CHANNELS_COUNT
)

// Statistics represents transfer statistics
type Statistics struct {
	channels []*StatChannel
}

// NewStatistics creates a new Statistics instance
func NewStatistics() *Statistics {
	s := &Statistics{
		channels: make([]*StatChannel, CHANNELS_COUNT),
	}
	for i := range s.channels {
		s.channels[i] = NewStatChannel()
	}
	return s
}

// Add adds another Statistics object to this one
func (s *Statistics) Add(stat *Statistics) *Statistics {
	for i := 0; i < int(CHANNELS_COUNT); i++ {
		s.channels[i].AddChannel(stat.channels[i])
	}
	return s
}

// SecondTick updates all channels with a second tick
func (s *Statistics) SecondTick(timeIntervalMS int64) {
	for _, sc := range s.channels {
		sc.SecondTick(timeIntervalMS)
	}
}

// Clear resets all statistics
func (s *Statistics) Clear() {
	for _, sc := range s.channels {
		sc.Clear()
	}
}

// ReceiveBytes records received bytes (protocol and payload)
func (s *Statistics) ReceiveBytes(protocolBytes, payloadBytes int64) {
	s.channels[DOWNLOAD_PROTOCOL].Add(protocolBytes)
	s.channels[DOWNLOAD_PAYLOAD].Add(payloadBytes)
}

// SendBytes records sent bytes (protocol and payload)
func (s *Statistics) SendBytes(protocolBytes, payloadBytes int64) {
	s.channels[UPLOAD_PROTOCOL].Add(protocolBytes)
	s.channels[UPLOAD_PAYLOAD].Add(payloadBytes)
}

// TotalPayloadDownload returns total payload bytes received
func (s *Statistics) TotalPayloadDownload() int64 {
	return s.channels[DOWNLOAD_PAYLOAD].Total()
}

// TotalProtocolDownload returns total protocol bytes received
func (s *Statistics) TotalProtocolDownload() int64 {
	return s.channels[DOWNLOAD_PROTOCOL].Total()
}

// TotalUpload returns total bytes uploaded (payload + protocol)
func (s *Statistics) TotalUpload() int64 {
	return s.channels[UPLOAD_PAYLOAD].Total() + s.channels[UPLOAD_PROTOCOL].Total()
}

// LastDownload returns last download bytes (payload + protocol)
func (s *Statistics) LastDownload() int64 {
	return s.channels[DOWNLOAD_PAYLOAD].Counter() + s.channels[DOWNLOAD_PROTOCOL].Counter()
}

// LastUpload returns last upload bytes (payload + protocol)
func (s *Statistics) LastUpload() int64 {
	return s.channels[UPLOAD_PAYLOAD].Counter() + s.channels[UPLOAD_PROTOCOL].Counter()
}

// DownloadRate returns total download rate (payload + protocol)
func (s *Statistics) DownloadRate() int64 {
	return s.channels[DOWNLOAD_PAYLOAD].Rate() + s.channels[DOWNLOAD_PROTOCOL].Rate()
}

// DownloadPayloadRate returns payload download rate
func (s *Statistics) DownloadPayloadRate() int64 {
	return s.channels[DOWNLOAD_PAYLOAD].Rate()
}

// UploadRate returns total upload rate (payload + protocol)
func (s *Statistics) UploadRate() int64 {
	return s.channels[UPLOAD_PAYLOAD].Rate() + s.channels[UPLOAD_PROTOCOL].Rate()
}

// UploadPayloadRate returns payload upload rate
func (s *Statistics) UploadPayloadRate() int64 {
	return s.channels[UPLOAD_PAYLOAD].Rate()
}

// LowPassUploadRate returns low pass upload rate
func (s *Statistics) LowPassUploadRate() int64 {
	return s.channels[UPLOAD_PAYLOAD].LowPassRate() + s.channels[UPLOAD_PROTOCOL].LowPassRate()
}

// LowPassDownloadRate returns low pass download rate
func (s *Statistics) LowPassDownloadRate() int64 {
	return s.channels[DOWNLOAD_PAYLOAD].LowPassRate() + s.channels[DOWNLOAD_PROTOCOL].LowPassRate()
}