package jed2k

import (
	"testing"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

func TestPeer(t *testing.T) {
	// Test basic peer creation
	endpoint := protocol.NewEndpointFromIPPort(0xC0A80101, 4661) // 192.168.1.1:4661
	peer := NewPeer(endpoint)
	
	if peer.GetEndpoint() != endpoint {
		t.Error("Peer endpoint should match")
	}
	
	if peer.IsConnectable() {
		t.Error("New peer should not be connectable by default")
	}
	
	if peer.HasConnection() {
		t.Error("New peer should not have connection")
	}
	
	// Test peer with flags
	peer2 := NewPeerWithFlags(endpoint, true, 4) // Use DHT flag value 4
	if !peer2.IsConnectable() {
		t.Error("Peer should be connectable when flag is set")
	}
	
	if peer2.GetSourceFlag() != 4 {
		t.Error("Source flag should match")
	}
	
	// Test peer comparison
	endpoint2 := protocol.NewEndpointFromIPPort(0xC0A80102, 4661) // 192.168.1.2:4661
	peer3 := NewPeer(endpoint2)
	
	if peer.Compare(peer3) >= 0 {
		t.Error("peer should be less than peer3")
	}
	
	if !peer.Equals(peer) {
		t.Error("Peer should equal itself")
	}
	
	if peer.Equals(peer3) {
		t.Error("Different peers should not be equal")
	}
}

func TestStatistics(t *testing.T) {
	stats := NewStatistics()
	
	// Test initial state
	if stats.TotalPayloadDownload() != 0 {
		t.Error("Initial payload download should be 0")
	}
	
	if stats.DownloadRate() != 0 {
		t.Error("Initial download rate should be 0")
	}
	
	// Test adding data
	stats.ReceiveBytes(100, 500) // 100 protocol, 500 payload
	if stats.TotalProtocolDownload() != 100 {
		t.Error("Protocol download should be 100")
	}
	
	if stats.TotalPayloadDownload() != 500 {
		t.Error("Payload download should be 500")
	}
	
	stats.SendBytes(50, 200) // 50 protocol, 200 payload
	if stats.TotalUpload() != 250 {
		t.Error("Total upload should be 250 (50+200)")
	}
	
	// Test second tick simulation
	stats.SecondTick(1000) // 1 second
	
	// Test statistics addition
	stats2 := NewStatistics()
	stats2.ReceiveBytes(25, 75)
	
	stats.Add(stats2)
	if stats.TotalProtocolDownload() != 125 { // 100 + 25
		t.Error("Combined protocol download should be 125")
	}
}

func TestSpeedMonitor(t *testing.T) {
	monitor := NewSpeedMonitor(5) // 5 samples
	
	// Test initial state
	if monitor.GetNumSamples() != 0 {
		t.Error("Initial samples should be 0")
	}
	
	if monitor.AverageSpeed() != INVALID_SPEED {
		t.Error("Average speed should be INVALID_SPEED when no samples")
	}
	
	// Add samples
	monitor.AddSample(100)
	monitor.AddSample(200)
	monitor.AddSample(300)
	
	if monitor.GetNumSamples() != 3 {
		t.Error("Should have 3 samples")
	}
	
	avgSpeed := monitor.AverageSpeed()
	expectedAvg := int64(200) // (100+200+300)/3
	if avgSpeed != expectedAvg {
		t.Errorf("Average speed should be %d, got %d", expectedAvg, avgSpeed)
	}
	
	// Test clear
	monitor.Clear()
	if monitor.GetNumSamples() != 0 {
		t.Error("After clear, samples should be 0")
	}
}

func TestPair(t *testing.T) {
	// Test string pair
	pair := NewPair("hello", "world")
	if pair.GetLeft() != "hello" {
		t.Error("Left value should be 'hello'")
	}
	
	if pair.GetRight() != "world" {
		t.Error("Right value should be 'world'")
	}
	
	// Test number pair
	numPair := NewPair(42, 3.14)
	if numPair.GetLeft() != 42 {
		t.Error("Left value should be 42")
	}
	
	if numPair.GetRight() != 3.14 {
		t.Error("Right value should be 3.14")
	}
	
	// Test equality (simplified string comparison)
	pair2 := NewPair("hello", "world")
	if !pair.Equals(pair2) {
		t.Error("Identical pairs should be equal")
	}
	
	pair3 := NewPair("hello", "mars")
	if pair.Equals(pair3) {
		t.Error("Different pairs should not be equal")
	}
	
	// Test string representation
	str := pair.String()
	if str != "Pair(left=hello, right=world)" {
		t.Errorf("String representation incorrect: %s", str)
	}
}

func TestTimeUtilities(t *testing.T) {
	// Test time functions
	start := CurrentTimeHiRes()
	UpdateCachedTime()
	cached := CurrentTime()
	
	if cached <= 0 {
		t.Error("Cached time should be positive")
	}
	
	if start <= 0 {
		t.Error("High resolution time should be positive")
	}
	
	// Test conversion functions
	sec := Seconds(5)
	if sec != 5000 {
		t.Error("5 seconds should be 5000 milliseconds")
	}
	
	min := Minutes(2)
	if min != 120000 {
		t.Error("2 minutes should be 120000 milliseconds")
	}
	
	hour := Hours(1)
	if hour != 3600000 {
		t.Error("1 hour should be 3600000 milliseconds")
	}
}