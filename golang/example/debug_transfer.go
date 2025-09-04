package main

import (
	"fmt"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

func main() {
	fmt.Println("===== Debugging Transfer State Issues =====")

	// Create session with debug info
	settings := jed2k.NewDefaultSettings()
	settings.PeerConnectionTimeout = 5
	session := jed2k.NewSession(settings)

	fmt.Printf("✓ Session created\n")

	// Start session
	if err := session.Start(); err != nil {
		fmt.Printf("❌ Failed to start session: %v\n", err)
		return
	}
	defer session.Stop()

	fmt.Printf("✓ Session started\n")

	// Add a transfer
	edkLink := "ed2k://|file|debug_test.avi|52428800|AABBCCDDEEFF1122334455667788990A|/"
	handle, err := session.AddTransferFromLink(edkLink, "./debug_downloads")
	if err != nil {
		fmt.Printf("❌ Failed to add transfer: %v\n", err)
		return
	}

	fmt.Printf("✓ Transfer added: %s\n", handle.GetName())

	// Wait a moment and then manually check the state
	fmt.Println("\n=== Checking Initial State ===")
	time.Sleep(100 * time.Millisecond)
	
	status := handle.GetStatus()
	fmt.Printf("Initial State: %s\n", status.State.String())
	fmt.Printf("Initial Peers: %d\n", len(handle.GetPeersInfo()))
	
	// Wait 3 seconds for session loop to process
	fmt.Println("\n=== Waiting for Session Processing (3 seconds) ===")
	time.Sleep(3 * time.Second)
	
	// Check state again
	status = handle.GetStatus()
	peers := handle.GetPeersInfo()
	fmt.Printf("After 3s - State: %s\n", status.State.String())
	fmt.Printf("After 3s - Peers: %d\n", len(peers))
	fmt.Printf("After 3s - Download Rate: %s/s\n", formatBytes(int64(status.DownloadRate)))
	fmt.Printf("After 3s - Progress: %.1f%%\n", status.Progress*100)
	
	// Print peer details
	fmt.Println("\n=== Peer Details ===")
	for i, peer := range peers {
		fmt.Printf("Peer %d: %s - Connected: %t - Rate: %s/s\n", 
			i+1, peer.Endpoint.String(), peer.Connected, formatBytes(int64(peer.DownloadRate)))
	}
	
	// Wait another 5 seconds
	fmt.Println("\n=== Waiting Additional 5 Seconds ===")
	time.Sleep(5 * time.Second)
	
	// Final check
	status = handle.GetStatus()
	peers = handle.GetPeersInfo()
	fmt.Printf("After 8s - State: %s\n", status.State.String())
	fmt.Printf("After 8s - Downloaded: %s / %s (%.1f%%)\n", 
		formatBytes(status.Downloaded), formatBytes(status.Size), status.Progress*100)
	fmt.Printf("After 8s - Rate: %s/s\n", formatBytes(int64(status.DownloadRate)))
	
	connected := 0
	for _, peer := range peers {
		if peer.Connected {
			connected++
		}
	}
	fmt.Printf("After 8s - Connected Peers: %d/%d\n", connected, len(peers))

	fmt.Printf("\n✓ Debug test completed\n")
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}