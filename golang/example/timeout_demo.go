package main

import (
	"fmt"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

func main() {
	fmt.Println("===== Testing Timeout and Retry Mechanisms =====")

	// Create session
	settings := jed2k.NewDefaultSettings()
	settings.PeerConnectionTimeout = 5 // 5 second timeout
	session := jed2k.NewSession(settings)

	fmt.Printf("Session settings:\n")
	fmt.Printf("  Peer Connection Timeout: %d seconds\n", settings.PeerConnectionTimeout)
	fmt.Printf("  Max Peer List Size: %d\n", settings.MaxPeerListSize)
	fmt.Printf("  Min Peer Reconnect Time: %d seconds\n", settings.MinPeerReconnectTime)

	// Start session
	if err := session.Start(); err != nil {
		fmt.Printf("❌ Failed to start session: %v\n", err)
		return
	}
	defer session.Stop()

	fmt.Println("✓ Session started")

	// Add a transfer
	edkLink := "ed2k://|file|timeout_test.avi|104857600|AABBCCDDEEFF1122334455667788990A|/"
	handle, err := session.AddTransferFromLink(edkLink, "./test_downloads")
	if err != nil {
		fmt.Printf("❌ Failed to add transfer: %v\n", err)
		return
	}

	fmt.Printf("✓ Transfer added: %s\n", handle.GetName())

	// Monitor transfer state for 30 seconds to see timeout and retry behavior
	fmt.Println("\n=== Monitoring Transfer State for Timeout/Retry Behavior ===")

	for i := 0; i < 30; i++ {
		status := handle.GetStatus()
		peers := handle.GetPeersInfo()
		
		connectedPeers := 0
		for _, peer := range peers {
			if peer.Connected {
				connectedPeers++
			}
		}

		fmt.Printf("Time: %2ds | State: %-11s | Progress: %5.1f%% | Download: %8s | Connected Peers: %d/%d\n",
			i, status.State.String(), status.Progress*100,
			formatBytes(int64(status.DownloadRate)), connectedPeers, len(peers))

		if status.State == jed2k.TransferStateCompleted {
			fmt.Println("✓ Transfer completed!")
			break
		}

		time.Sleep(1 * time.Second)
	}

	// Show final statistics
	fmt.Println("\n=== Final Statistics ===")
	finalStatus := handle.GetStatus()
	sessionStats := session.GetSessionStats()

	fmt.Printf("Transfer Status:\n")
	fmt.Printf("  State: %s\n", finalStatus.State.String())
	fmt.Printf("  Progress: %.1f%%\n", finalStatus.Progress*100)
	fmt.Printf("  Downloaded: %s / %s\n", formatBytes(finalStatus.Downloaded), formatBytes(finalStatus.Size))
	fmt.Printf("  Download Rate: %s/s\n", formatBytes(int64(finalStatus.DownloadRate)))
	fmt.Printf("  Connected Peers: %d\n", finalStatus.ConnectedPeers)

	fmt.Printf("\nSession Stats:\n")
	fmt.Printf("  Total Transfers: %d\n", sessionStats.TotalTransfers)
	fmt.Printf("  Active Transfers: %d\n", sessionStats.ActiveTransfers)
	fmt.Printf("  Connected Peers: %d\n", sessionStats.ConnectedPeers)

	fmt.Println("\n✓ Test completed - timeout and retry mechanisms demonstrated")
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