package main

import (
	"fmt"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

func main() {
	fmt.Println("===== Testing Timeout and Retry Mechanisms Manually =====")

	// Create session
	settings := jed2k.NewDefaultSettings()
	settings.PeerConnectionTimeout = 5
	session := jed2k.NewSession(settings)
	
	if err := session.Start(); err != nil {
		fmt.Printf("❌ Failed to start session: %v\n", err)
		return
	}
	defer session.Stop()
	
	// Add a transfer
	edkLink := "ed2k://|file|manual_test.avi|1048576|AABBCCDDEEFF1122334455667788990A|/"
	handle, err := session.AddTransferFromLink(edkLink, "./downloads")
	if err != nil {
		fmt.Printf("❌ Failed to add transfer: %v\n", err)
		return
	}

	fmt.Printf("✓ Transfer added: %s\n", handle.GetName())
	
	// Get the underlying transfer object (we need to access it directly)
	transfers := session.GetTransfers()
	if len(transfers) == 0 {
		fmt.Printf("❌ No transfers found\n")
		return
	}
	
	transfer := transfers[0].(*jed2k.TransferHandle) // Cast to access internal methods
	
	fmt.Printf("✓ Got transfer handle\n")

	// Check initial status
	status := transfer.GetStatus()
	fmt.Printf("Initial State: %s\n", status.State.String())
	fmt.Printf("Initial Peers: %d\n", len(transfer.GetPeersInfo()))
	
	// Manually simulate what the session loop would do
	fmt.Println("\n=== Manual Session Update Simulation ===")
	
	// Simulate the update process multiple times
	for i := 0; i < 5; i++ {
		fmt.Printf("--- Update %d ---\n", i+1)
		
		// This would be called by session.updateTransfers()
		status = transfer.GetStatus()
		peers := transfer.GetPeersInfo()
		
		fmt.Printf("State: %s, Peers: %d\n", status.State.String(), len(peers))
		
		if status.State == jed2k.TransferStateDownloading {
			fmt.Printf("✓ Transfer is now downloading!\n")
			fmt.Printf("Download Rate: %.1f KB/s\n", status.DownloadRate/1024)
			fmt.Printf("Progress: %.1f%%\n", status.Progress*100)
			
			// Count connected peers
			connected := 0
			for _, peer := range peers {
				if peer.Connected {
					connected++
				}
			}
			fmt.Printf("Connected Peers: %d/%d\n", connected, len(peers))
		}
		
		if status.State == jed2k.TransferStateCompleted {
			fmt.Printf("✓ Transfer completed!\n")
			break
		}
	}
	
	// Final status
	finalStatus := transfer.GetStatus()
	fmt.Printf("\n=== Final Status ===\n")
	fmt.Printf("State: %s\n", finalStatus.State.String())
	fmt.Printf("Progress: %.1f%%\n", finalStatus.Progress*100)
	fmt.Printf("Downloaded: %d / %d bytes\n", finalStatus.Downloaded, finalStatus.Size)

	fmt.Printf("✓ Manual test completed\n")
}