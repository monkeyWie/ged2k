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
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// ExampleClient demonstrates comprehensive usage of the ed2k client
type ExampleClient struct {
	session           *jed2k.Session
	running           bool
	lastSearchResults []*protocol.SearchEntry
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
		fmt.Printf("\nDiscovered Peers (%d total):\n", len(peers))
		fmt.Printf("%-20s %-15s %-10s %-10s %-15s\n", "Endpoint", "Client", "Downloaded", "Uploaded", "Status")
		fmt.Println(strings.Repeat("-", 75))
		
		for i, peer := range peers {
			if i >= 20 { // Limit to first 20 peers for readability
				fmt.Printf("... and %d more peers (use 'peers <index>' to see all)\n", len(peers)-20)
				break
			}
			
			status := "disconnected"
			if peer.Connected {
				status = "connected"
			}
			
			fmt.Printf("%-20s %-15s %-10s %-10s %-15s\n", 
				peer.Endpoint.String(), 
				truncateString(peer.ClientName, 15),
				formatBytes(peer.Downloaded), 
				formatBytes(peer.Uploaded),
				status)
		}
	} else {
		fmt.Printf("\nNo peers discovered yet. Peer discovery status:\n")
		fmt.Printf("  - Server queries: in progress\n")
		fmt.Printf("  - DHT queries: not yet implemented\n")
		fmt.Printf("  - Peer exchange: not yet implemented\n")
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

// ShowAllPeers shows all discovered peers for a specific transfer
func (ec *ExampleClient) ShowAllPeers(index int) error {
	handles := ec.session.GetTransfers()
	if index < 1 || index > len(handles) {
		return fmt.Errorf("invalid transfer index: %d", index)
	}
	
	handle := handles[index-1]
	status := handle.GetStatus()
	peers := handle.GetPeersInfo()
	
	fmt.Printf("\n=== All Discovered Peers for Transfer ===\n")
	fmt.Printf("Transfer: %s\n", status.Name)
	fmt.Printf("Hash: %s\n", status.Hash.String())
	fmt.Printf("Total Peers: %d\n", len(peers))
	
	if len(peers) == 0 {
		fmt.Printf("\nNo peers discovered yet. This could mean:\n")
		fmt.Printf("  1. Server connections are still establishing\n")
		fmt.Printf("  2. Servers have no sources for this file\n") 
		fmt.Printf("  3. File is very rare or unavailable\n")
		fmt.Printf("\nCheck server connection status and try requesting sources again.\n")
		return nil
	}
	
	fmt.Printf("\n%-25s %-20s %-12s %-12s %-10s %-15s\n", 
		"Endpoint", "Client", "Downloaded", "Uploaded", "Rate", "Status")
	fmt.Println(strings.Repeat("-", 100))
	
	for i, peer := range peers {
		status := "disconnected"
		if peer.Connected {
			status = "connected"
		}
		
		rate := ""
		if peer.DownloadRate > 0 || peer.UploadRate > 0 {
			rate = fmt.Sprintf("↓%s ↑%s", 
				formatRate(peer.DownloadRate), 
				formatRate(peer.UploadRate))
		}
		
		fmt.Printf("%-25s %-20s %-12s %-12s %-10s %-15s\n",
			peer.Endpoint.String(),
			truncateString(peer.ClientName, 20),
			formatBytes(peer.Downloaded),
			formatBytes(peer.Uploaded),
			truncateString(rate, 10),
			status)
			
		// Group by every 10 for readability
		if (i+1) % 10 == 0 && i+1 < len(peers) {
			fmt.Printf("\n--- Showing peers %d-%d of %d (press Enter to continue) ---\n", i-8, i+1, len(peers))
			fmt.Scanln() // Wait for user input
		}
	}
	
	return nil
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

// SearchFiles performs a search for files on the ed2k network
func (ec *ExampleClient) SearchFiles(query, fileType string, minSize, maxSize uint64) error {
	fmt.Printf("\n=== Searching Files ===\n")
	fmt.Printf("Query: %s\n", query)
	if fileType != "" {
		fmt.Printf("File Type: %s\n", fileType)
	}
	if minSize > 0 {
		fmt.Printf("Min Size: %.2f MB\n", float64(minSize)/(1024*1024))
	}
	if maxSize > 0 {
		fmt.Printf("Max Size: %.2f MB\n", float64(maxSize)/(1024*1024))
	}
	
	// Perform search
	if err := ec.session.SearchSimple(query, fileType, minSize, maxSize); err != nil {
		return fmt.Errorf("failed to perform search: %v", err)
	}
	
	fmt.Printf("✓ Search request sent to all connected servers\n")
	fmt.Println("Use 'results' command to view search results when available")
	
	return nil
}

// ShowSearchResults displays the last search results
func (ec *ExampleClient) ShowSearchResults() {
	fmt.Printf("\n=== Search Results ===\n")
	
	if len(ec.lastSearchResults) == 0 {
		fmt.Println("No search results available.")
		fmt.Println("Use 'search <query>' to perform a search first.")
		return
	}
	
	fmt.Printf("Found %d result(s):\n\n", len(ec.lastSearchResults))
	fmt.Printf("%-3s %-40s %-10s %-8s %-12s\n", "#", "Name", "Size", "Sources", "Hash")
	fmt.Println(strings.Repeat("-", 80))
	
	for i, entry := range ec.lastSearchResults {
		name := entry.GetFilename()
		if len(name) > 38 {
			name = name[:35] + "..."
		}
		
		sizeStr := formatBytes(int64(entry.GetFilesize()))
		
		fmt.Printf("%-3d %-40s %-10s %-8d %-12s\n",
			i+1, name, sizeStr, entry.GetSources(), entry.GetHash().String()[:12]+"...")
	}
	
	fmt.Println("\nUse 'load <index>' to add a result to downloads")
}

// LoadSearchResult adds a search result to downloads
func (ec *ExampleClient) LoadSearchResult(index int) error {
	if len(ec.lastSearchResults) == 0 {
		return fmt.Errorf("no search results available")
	}
	
	if index < 1 || index > len(ec.lastSearchResults) {
		return fmt.Errorf("invalid result index: %d (available: 1-%d)", index, len(ec.lastSearchResults))
	}
	
	entry := ec.lastSearchResults[index-1]
	
	fmt.Printf("\n=== Adding Search Result ===\n")
	fmt.Printf("Name: %s\n", entry.GetFilename())
	fmt.Printf("Size: %.2f MB\n", float64(entry.GetFilesize())/(1024*1024))
	fmt.Printf("Hash: %s\n", entry.GetHash().String())
	fmt.Printf("Sources: %d\n", entry.GetSources())
	
	// Convert to ed2k link and add as transfer
	link := entry.ToED2KLink()
	handle, err := ec.session.AddTransferFromLink(link, "./downloads")
	if err != nil {
		return fmt.Errorf("failed to add transfer: %v", err)
	}
	
	fmt.Printf("✓ Transfer added successfully\n")
	fmt.Printf("  Transfer Hash: %s\n", handle.GetHash().String())
	
	return nil
}

// ProcessSearchAlerts processes incoming search alerts
func (ec *ExampleClient) ProcessSearchAlerts() {
	// This would be called periodically to process search results
	// In a real implementation, you'd have an alert processing loop
	// For now, we'll simulate getting search results
}

// UpdateSearchResults updates the stored search results from alerts
func (ec *ExampleClient) UpdateSearchResults(results []*protocol.SearchEntry) {
	ec.lastSearchResults = results
	fmt.Printf("\n[SEARCH UPDATE] Received %d new search results\n", len(results))
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
	fmt.Println("  peers <index>                   - Show all discovered peers for transfer")
	fmt.Println("  pause <index>                   - Pause a transfer")
	fmt.Println("  resume <index>                  - Resume a transfer")
	fmt.Println("  pauseall                        - Pause all transfers")
	fmt.Println("  resumeall                       - Resume all transfers")
	fmt.Println("  search <query> [filetype] [minsize] [maxsize] - Search for files")
	fmt.Println("  results                         - Show last search results")
	fmt.Println("  load <result_index>             - Add search result to downloads")
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
			
		case "peers":
			if len(parts) < 2 {
				fmt.Println("Usage: peers <index>")
				continue
			}
			index, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid index")
				continue
			}
			if err := ec.ShowAllPeers(index); err != nil {
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
			
		case "search":
			if len(parts) < 2 {
				fmt.Println("Usage: search <query> [filetype] [minsize_mb] [maxsize_mb]")
				fmt.Println("Examples:")
				fmt.Println("  search \"game of thrones\"")
				fmt.Println("  search music Audio 0 500")
				fmt.Println("  search documentary Video")
				continue
			}
			
			query := parts[1]
			fileType := ""
			var minSize, maxSize uint64
			
			if len(parts) > 2 {
				fileType = parts[2]
			}
			if len(parts) > 3 {
				if minMB, err := strconv.Atoi(parts[3]); err == nil && minMB > 0 {
					minSize = uint64(minMB) * 1024 * 1024
				}
			}
			if len(parts) > 4 {
				if maxMB, err := strconv.Atoi(parts[4]); err == nil && maxMB > 0 {
					maxSize = uint64(maxMB) * 1024 * 1024
				}
			}
			
			if err := ec.SearchFiles(query, fileType, minSize, maxSize); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			
		case "results":
			ec.ShowSearchResults()
			
		case "load":
			if len(parts) < 2 {
				fmt.Println("Usage: load <result_index>")
				fmt.Println("Use 'results' to see available search results")
				continue
			}
			index, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid result index")
				continue
			}
			if err := ec.LoadSearchResult(index); err != nil {
				fmt.Printf("Error: %v\n", err)
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
			fmt.Println("  peers <index>                   - Show all discovered peers for transfer")
			fmt.Println("  pause <index>                   - Pause a transfer")
			fmt.Println("  resume <index>                  - Resume a transfer")
			fmt.Println("  pauseall                        - Pause all transfers")
			fmt.Println("  resumeall                       - Resume all transfers")
			fmt.Println("  search <query> [filetype] [minsize] [maxsize] - Search for files")
			fmt.Println("  results                         - Show last search results")
			fmt.Println("  load <result_index>             - Add search result to downloads")
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

// formatBytes formats bytes to human readable format
func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
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