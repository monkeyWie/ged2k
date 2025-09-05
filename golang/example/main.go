package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"
)

// ExampleClient demonstrates comprehensive usage of the ed2k client
type ExampleClient struct {
	session *jed2k.Session
	running bool
}

// NewExampleClient creates a new example client
func NewExampleClient() *ExampleClient {
	return &ExampleClient{}
}

// Initialize sets up the ed2k client with configuration
func (ec *ExampleClient) Initialize() error {
	fmt.Println("=== Initializing ED2K Client ===")
	
	// Create custom settings
	settings := jed2k.NewDefaultSettings()
	settings.ListenPort = 4661
	settings.UDPPort = 4662
	settings.IncomingDirectory = "./downloads"
	settings.ModName = "ged2k-example"
	settings.ClientName = "ged2k Go Client"
	
	fmt.Printf("Client Settings:\n")
	fmt.Printf("  Listen Port: %d\n", settings.ListenPort)
	fmt.Printf("  UDP Port: %d\n", settings.UDPPort)
	fmt.Printf("  Download Directory: %s\n", settings.IncomingDirectory)
	
	// Create session with disk-based persistence
	fmt.Println("\n=== Configuring Persistence ===")
	fmt.Println("Available persistence options:")
	fmt.Println("  1. Memory-based (data lost on restart)")  
	fmt.Println("  2. Disk-based (data persisted to ./resume_data)")
	fmt.Println("Using disk-based persistence for this example...")
	ec.session = jed2k.NewSessionWithDiskPersistence(settings, "./resume_data")
	
	// Example showing memory-based persistence (commented out):
	// ec.session = jed2k.NewSessionWithDefaults(settings)
	
	// Configure server list
	fmt.Println("\n=== Configuring Server List ===")
	
	// Add manual servers
	fmt.Println("Adding manual servers...")
	// Note: In real implementation, these would be actual server endpoints
	
	// Load from server.met URL (demonstrates enhanced connectivity)
	fmt.Println("Loading servers from server.met...")
	if err := ec.session.LoadServerList("http://upd.emule-security.org/server.met"); err != nil {
		fmt.Printf("Warning: Failed to load server.met: %v\n", err)
	} else {
		fmt.Println("✓ Successfully loaded server.met")
	}
	
	// Load from nodes.dat URL (demonstrates Kademlia/DHT support)
	fmt.Println("Loading Kademlia nodes from nodes.dat...")
	if err := ec.session.LoadNodesList("http://upd.emule-security.org/nodes.dat"); err != nil {
		fmt.Printf("Warning: Failed to load nodes.dat: %v\n", err)
	} else {
		fmt.Println("✓ Successfully loaded nodes.dat")
	}
	
	// Start the session
	fmt.Println("\n=== Starting Session ===")
	if err := ec.session.Start(); err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	
	ec.running = true
	fmt.Println("✓ Session started successfully")
	
	return nil
}

// AddDownload demonstrates adding ed2k:// file downloads
func (ec *ExampleClient) AddDownload(linkStr, downloadDir string) error {
	fmt.Printf("\n=== Adding Download ===\n")
	fmt.Printf("Link: %s\n", linkStr)
	fmt.Printf("Download Directory: %s\n", downloadDir)
	
	// Parse the ed2k link
	link, err := jed2k.ParseEMuleLink(linkStr)
	if err != nil {
		return fmt.Errorf("failed to parse ed2k link: %v", err)
	}
	
	fmt.Printf("Parsed Link:\n")
	fmt.Printf("  Type: %s\n", map[jed2k.LinkType]string{
		jed2k.LinkTypeFile: "File",
		jed2k.LinkTypeServer: "Server", 
		jed2k.LinkTypeSearch: "Search",
	}[link.LinkType])
	fmt.Printf("  Name: %s\n", link.Name)
	fmt.Printf("  Size: %d bytes (%.2f MB)\n", link.Size, float64(link.Size)/(1024*1024))
	fmt.Printf("  Hash: %s\n", link.Hash.String())
	
	if !link.IsFileLink() {
		return fmt.Errorf("only file links are supported")
	}
	
	// Add the transfer
	handle, err := ec.session.AddTransferFromLink(linkStr, downloadDir)
	if err != nil {
		return fmt.Errorf("failed to add transfer: %v", err)
	}
	
	fmt.Printf("✓ Transfer added successfully\n")
	fmt.Printf("  Transfer Hash: %s\n", handle.GetHash().String())
	fmt.Printf("  Transfer Size: %d bytes\n", handle.GetSize())
	fmt.Printf("  Transfer Name: %s\n", handle.GetName())
	
	// Small delay to allow state transition to occur
	time.Sleep(200 * time.Millisecond)
	
	return nil
}

// ListTransfers shows all current transfers with their status
func (ec *ExampleClient) ListTransfers() {
	fmt.Println("\n=== Current Transfers ===")
	
	handles := ec.session.GetTransfers()
	if len(handles) == 0 {
		fmt.Println("No active transfers")
		return
	}
	
	fmt.Printf("%-3s %-32s %-20s %-10s %-8s %-8s %-12s %-12s\n", 
		"#", "Hash", "Name", "State", "Progress", "Down", "Up", "Peers")
	fmt.Println(strings.Repeat("-", 120))
	
	for i, handle := range handles {
		status := handle.GetStatus()
		
		// Truncate name if too long
		name := status.Name
		if len(name) > 18 {
			name = name[:15] + "..."
		}
		
		// Format progress
		progressStr := fmt.Sprintf("%.1f%%", status.Progress*100)
		
		// Format rates
		downRate := formatRate(status.DownloadRate)
		upRate := formatRate(status.UploadRate)
		
		// Format peers
		peersStr := fmt.Sprintf("%d/%d", status.ConnectedPeers, status.TotalPeers)
		
		fmt.Printf("%-3d %-32s %-20s %-10s %-8s %-8s %-12s %-12s\n",
			i+1, status.Hash.String()[:8]+"...", name, status.State.String(),
			progressStr, downRate, upRate, peersStr)
	}
}

// PauseTransfer pauses a specific transfer by index
func (ec *ExampleClient) PauseTransfer(index int) error {
	handles := ec.session.GetTransfers()
	if index < 1 || index > len(handles) {
		return fmt.Errorf("invalid transfer index: %d", index)
	}
	
	handle := handles[index-1]
	if err := handle.Pause(); err != nil {
		return fmt.Errorf("failed to pause transfer: %v", err)
	}
	
	fmt.Printf("✓ Paused transfer: %s\n", handle.GetName())
	return nil
}

// ResumeTransfer resumes a specific transfer by index
func (ec *ExampleClient) ResumeTransfer(index int) error {
	handles := ec.session.GetTransfers()
	if index < 1 || index > len(handles) {
		return fmt.Errorf("invalid transfer index: %d", index)
	}
	
	handle := handles[index-1]
	if err := handle.Resume(); err != nil {
		return fmt.Errorf("failed to resume transfer: %v", err)
	}
	
	fmt.Printf("✓ Resumed transfer: %s\n", handle.GetName())
	return nil
}

// ShowTransferDetails shows detailed information about a specific transfer
func (ec *ExampleClient) ShowTransferDetails(index int) error {
	handles := ec.session.GetTransfers()
	if index < 1 || index > len(handles) {
		return fmt.Errorf("invalid transfer index: %d", index)
	}
	
	handle := handles[index-1]
	status := handle.GetStatus()
	peers := handle.GetPeersInfo()
	
	fmt.Printf("\n=== Transfer Details ===\n")
	fmt.Printf("Name: %s\n", status.Name)
	fmt.Printf("Hash: %s\n", status.Hash.String())
	fmt.Printf("Size: %d bytes (%.2f MB)\n", status.Size, float64(status.Size)/(1024*1024))
	fmt.Printf("Downloaded: %d bytes (%.2f MB)\n", status.Downloaded, float64(status.Downloaded)/(1024*1024))
	fmt.Printf("Uploaded: %d bytes (%.2f MB)\n", status.Uploaded, float64(status.Uploaded)/(1024*1024))
	fmt.Printf("Progress: %.2f%%\n", status.Progress*100)
	fmt.Printf("State: %s\n", status.State.String())
	fmt.Printf("Download Rate: %s\n", formatRate(status.DownloadRate))
	fmt.Printf("Upload Rate: %s\n", formatRate(status.UploadRate))
	fmt.Printf("Connected Peers: %d\n", status.ConnectedPeers)
	fmt.Printf("Total Peers: %d\n", status.TotalPeers)
	fmt.Printf("Seeds: %d\n", status.Seeds)
	fmt.Printf("Download Directory: %s\n", status.DownloadDirectory)
	
	if status.ETA > 0 {
		fmt.Printf("ETA: %s\n", status.ETA.String())
	}
	
	if status.ErrorMessage != "" {
		fmt.Printf("Error: %s\n", status.ErrorMessage)
	}
	
	if len(peers) > 0 {
		fmt.Printf("\nConnected Peers:\n")
		for i, peer := range peers {
			if i >= 10 { // Limit to first 10 peers
				fmt.Printf("... and %d more peers\n", len(peers)-10)
				break
			}
			fmt.Printf("  %s - %s (↓ %s, ↑ %s)\n", 
				peer.Endpoint.String(), peer.ClientName,
				formatRate(peer.DownloadRate), formatRate(peer.UploadRate))
		}
	}
	
	return nil
}

// ShowSessionStats displays overall session statistics
func (ec *ExampleClient) ShowSessionStats() {
	fmt.Println("\n=== Session Statistics ===")
	
	stats := ec.session.GetSessionStats()
	
	fmt.Printf("Transfers:\n")
	fmt.Printf("  Total: %d\n", stats.TotalTransfers)
	fmt.Printf("  Active: %d\n", stats.ActiveTransfers)
	fmt.Printf("  Paused: %d\n", stats.PausedTransfers)
	fmt.Printf("  Completed: %d\n", stats.CompletedTransfers)
	
	fmt.Printf("Traffic:\n")
	fmt.Printf("  Downloaded: %.2f MB\n", float64(stats.TotalDownloaded)/(1024*1024))
	fmt.Printf("  Uploaded: %.2f MB\n", float64(stats.TotalUploaded)/(1024*1024))
	fmt.Printf("  Download Rate: %s\n", formatRate(stats.GlobalDownloadRate))
	fmt.Printf("  Upload Rate: %s\n", formatRate(stats.GlobalUploadRate))
	
	fmt.Printf("Network:\n")
	fmt.Printf("  Connected Peers: %d\n", stats.ConnectedPeers)
	fmt.Printf("  Known Servers: %d\n", stats.KnownServers)
	fmt.Printf("  Known Nodes: %d\n", stats.KnownNodes)
}

// DemonstratePersistence shows different persistence implementations
func (ec *ExampleClient) DemonstratePersistence() {
	fmt.Println("\n=== Persistence Demonstration ===")
	
	// Create different persistence managers
	memoryPersistence := jed2k.NewPersistenceManager(jed2k.NewMemoryResumeData())
	diskPersistence := jed2k.NewPersistenceManager(jed2k.NewDiskResumeData("./example_resume"))
	
	fmt.Println("Available persistence options:")
	fmt.Println("1. Memory-based persistence (fast, not persistent across restarts)")
	fmt.Println("2. Disk-based persistence (slower, persistent across restarts)")
	fmt.Printf("Current configuration: %s\n", 
		map[bool]string{true: "Disk-based", false: "Memory-based"}[ec.session != nil])
	
	// In a real implementation, you could switch persistence managers
	_ = memoryPersistence
	_ = diskPersistence
}

// Shutdown gracefully shuts down the client
func (ec *ExampleClient) Shutdown() error {
	if !ec.running {
		return nil
	}
	
	fmt.Println("\n=== Shutting Down ===")
	fmt.Println("Saving transfer resume data...")
	
	if err := ec.session.Stop(); err != nil {
		return fmt.Errorf("failed to stop session: %v", err)
	}
	
	ec.running = false
	fmt.Println("✓ Session stopped successfully")
	return nil
}

// RunInteractiveMode starts an interactive command-line interface
func (ec *ExampleClient) RunInteractiveMode() {
	fmt.Println("\n=== Interactive Mode ===")
	fmt.Println("Available commands:")
	fmt.Println("  add <ed2k_link> [download_dir]  - Add a download")
	fmt.Println("  list                            - List all transfers")
	fmt.Println("  details <index>                 - Show transfer details")
	fmt.Println("  pause <index>                   - Pause a transfer")
	fmt.Println("  resume <index>                  - Resume a transfer")
	fmt.Println("  pauseall                        - Pause all transfers")
	fmt.Println("  resumeall                       - Resume all transfers")
	fmt.Println("  stats                           - Show session statistics")
	fmt.Println("  persistence                     - Show persistence options")
	fmt.Println("  help                            - Show this help")
	fmt.Println("  quit                            - Quit the application")
	fmt.Println()
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for ec.running {
		fmt.Print("ged2k> ")
		
		if !scanner.Scan() {
			break
		}
		
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		command := parts[0]
		
		switch command {
		case "add":
			if len(parts) < 2 {
				fmt.Println("Usage: add <ed2k_link> [download_dir]")
				continue
			}
			downloadDir := "./downloads"
			if len(parts) > 2 {
				downloadDir = parts[2]
			}
			// Remove quotes from ed2k link if present
			link := parts[1]
			if strings.HasPrefix(link, "\"") && strings.HasSuffix(link, "\"") {
				link = strings.Trim(link, "\"")
			}
			if err := ec.AddDownload(link, downloadDir); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			
		case "list":
			ec.ListTransfers()
			
		case "details":
			if len(parts) < 2 {
				fmt.Println("Usage: details <index>")
				continue
			}
			index, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid index")
				continue
			}
			if err := ec.ShowTransferDetails(index); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			
		case "pause":
			if len(parts) < 2 {
				fmt.Println("Usage: pause <index>")
				continue
			}
			index, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid index")
				continue
			}
			if err := ec.PauseTransfer(index); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			
		case "resume":
			if len(parts) < 2 {
				fmt.Println("Usage: resume <index>")
				continue
			}
			index, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid index")
				continue
			}
			if err := ec.ResumeTransfer(index); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			
		case "pauseall":
			if err := ec.session.PauseAll(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("✓ All transfers paused")
			}
			
		case "resumeall":
			if err := ec.session.ResumeAll(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("✓ All transfers resumed")
			}
			
		case "stats":
			ec.ShowSessionStats()
			
		case "persistence":
			ec.DemonstratePersistence()
			
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  add <ed2k_link> [download_dir]  - Add a download")
			fmt.Println("  list                            - List all transfers")
			fmt.Println("  details <index>                 - Show transfer details")
			fmt.Println("  pause <index>                   - Pause a transfer")
			fmt.Println("  resume <index>                  - Resume a transfer")
			fmt.Println("  pauseall                        - Pause all transfers")
			fmt.Println("  resumeall                       - Resume all transfers")
			fmt.Println("  stats                           - Show session statistics")
			fmt.Println("  persistence                     - Show persistence options")
			fmt.Println("  help                            - Show this help")
			fmt.Println("  quit                            - Quit the application")
			
		case "quit", "exit":
			fmt.Println("Exiting...")
			return
			
		default:
			fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", command)
		}
	}
}

// formatRate formats a transfer rate in bytes/second to human readable format
func formatRate(rate float64) string {
	if rate < 1024 {
		return fmt.Sprintf("%.0f B/s", rate)
	} else if rate < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", rate/1024)
	} else {
		return fmt.Sprintf("%.1f MB/s", rate/(1024*1024))
	}
}

// DemoExamples demonstrates the client with example ed2k links
func (ec *ExampleClient) DemoExamples() {
	fmt.Println("\n=== Running Demo Examples ===")
	
	// Example ed2k links (these are examples, replace with real links)
	examples := []struct {
		link string
		dir  string
	}{
		{
			link: "ed2k://|file|example_file1.pdf|1048576|31D6CFE0D16AE931B73C59D7E0C089C0|/",
			dir:  "./downloads/documents",
		},
		{
			link: "ed2k://|file|example_video.mp4|104857600|F4E6C8A2C1B7E1D4A8F5C3E9D2B1A6C8|/",
			dir:  "./downloads/videos",
		},
		{
			link: "ed2k://|file|example_archive.zip|52428800|A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6|/",
			dir:  "./downloads/archives",
		},
	}
	
	for i, example := range examples {
		fmt.Printf("\n--- Example %d ---\n", i+1)
		if err := ec.AddDownload(example.link, example.dir); err != nil {
			fmt.Printf("Failed to add example %d: %v\n", i+1, err)
		}
		time.Sleep(500 * time.Millisecond) // Small delay for demonstration
	}
	
	// Show transfers
	time.Sleep(1 * time.Second)
	ec.ListTransfers()
	
	// Demonstrate pause/resume
	fmt.Println("\n--- Pause/Resume Demo ---")
	if len(ec.session.GetTransfers()) > 0 {
		fmt.Println("Pausing first transfer...")
		ec.PauseTransfer(1)
		
		time.Sleep(2 * time.Second)
		ec.ListTransfers()
		
		fmt.Println("Resuming first transfer...")
		ec.ResumeTransfer(1)
		
		time.Sleep(2 * time.Second)
		ec.ListTransfers()
	}
	
	// Show statistics
	ec.ShowSessionStats()
}

func main() {
	fmt.Println("===== ED2K Go Client Example =====")
	fmt.Println("This example demonstrates a comprehensive ed2k client implementation")
	fmt.Println()
	
	// Create and initialize client
	client := NewExampleClient()
	
	if err := client.Initialize(); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	
	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\nReceived shutdown signal...")
		client.Shutdown()
		os.Exit(0)
	}()
	
	// Check command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "demo":
			// Run automated demo
			client.DemoExamples()
			
			// Wait a bit to see the transfers
			fmt.Println("\nDemo completed. Waiting 10 seconds before shutdown...")
			time.Sleep(10 * time.Second)
			
		case "interactive":
			// Run interactive mode
			client.RunInteractiveMode()
			
		default:
			fmt.Printf("Unknown mode: %s\n", os.Args[1])
			fmt.Println("Available modes: demo, interactive")
			os.Exit(1)
		}
	} else {
		// Default: run interactive mode
		client.RunInteractiveMode()
	}
	
	// Shutdown
	if err := client.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	
	fmt.Println("Good bye!")
}