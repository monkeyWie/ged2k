package main

import (
	"fmt"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

func main() {
	fmt.Println("===== Simple Session Test =====")

	// Create and start session
	session := jed2k.NewSession(nil)
	fmt.Printf("✓ Session created\n")
	
	if err := session.Start(); err != nil {
		fmt.Printf("❌ Failed to start session: %v\n", err)
		return
	}
	defer session.Stop()
	
	fmt.Printf("✓ Session started\n")

	// Add a transfer
	edkLink := "ed2k://|file|test.avi|1048576|AABBCCDDEEFF1122334455667788990A|/"
	handle, err := session.AddTransferFromLink(edkLink, "./downloads")
	if err != nil {
		fmt.Printf("❌ Failed to add transfer: %v\n", err)
		return
	}

	fmt.Printf("✓ Transfer added: %s\n", handle.GetName())
	
	// Check initial status
	status := handle.GetStatus()
	fmt.Printf("Initial status: %s\n", status.State.String())
	
	fmt.Printf("✓ Test completed without hanging\n")
}