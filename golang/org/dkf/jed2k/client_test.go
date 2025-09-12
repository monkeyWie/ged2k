package jed2k

import (
	"testing"
	"path/filepath"
	"os"
)

func TestEMuleLinkParsing(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		wantErr  bool
		wantType LinkType
		wantName string
		wantSize int64
	}{
		{
			name:     "valid file link",
			link:     "ed2k://|file|test.pdf|1048576|31D6CFE0D16AE931B73C59D7E0C089C0|/",
			wantErr:  false,
			wantType: LinkTypeFile,
			wantName: "test.pdf",
			wantSize: 1048576,
		},
		{
			name:     "invalid prefix",
			link:     "http://example.com",
			wantErr:  true,
		},
		{
			name:     "invalid format",
			link:     "ed2k://file/test.pdf",
			wantErr:  true,
		},
		{
			name:     "insufficient parameters",
			link:     "ed2k://|file|test.pdf|/",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEMuleLink(tt.link)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseEMuleLink() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseEMuleLink() unexpected error: %v", err)
				return
			}

			if result.LinkType != tt.wantType {
				t.Errorf("ParseEMuleLink() type = %v, want %v", result.LinkType, tt.wantType)
			}

			if result.Name != tt.wantName {
				t.Errorf("ParseEMuleLink() name = %v, want %v", result.Name, tt.wantName)
			}

			if result.Size != tt.wantSize {
				t.Errorf("ParseEMuleLink() size = %v, want %v", result.Size, tt.wantSize)
			}
		})
	}
}

func TestSessionBasics(t *testing.T) {
	settings := NewDefaultSettings()
	settings.IncomingDirectory = filepath.Join(os.TempDir(), "test_downloads")
	settings.ResumeDataDirectory = filepath.Join(os.TempDir(), "test_resume")
	
	session := NewSession(settings)
	
	if session == nil {
		t.Fatal("NewSession() returned nil")
	}
	
	err := session.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	
	if !session.IsRunning() {
		t.Error("Session should be running after Start()")
	}
	
	err = session.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
	
	if session.IsRunning() {
		t.Error("Session should not be running after Stop()")
	}
}

func TestTransferOperations(t *testing.T) {
	settings := NewDefaultSettings()
	settings.IncomingDirectory = filepath.Join(os.TempDir(), "test_downloads")
	settings.ResumeDataDirectory = filepath.Join(os.TempDir(), "test_resume")
	
	session := NewSession(settings)
	err := session.Start()
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer session.Stop()
	
	// Test adding transfer
	link := "ed2k://|file|test.pdf|1048576|31D6CFE0D16AE931B73C59D7E0C089C0|/"
	handle, err := session.AddTransferFromLink(link, settings.IncomingDirectory)
	if err != nil {
		t.Fatalf("AddTransferFromLink() failed: %v", err)
	}
	
	if !handle.IsValid() {
		t.Error("Transfer handle should be valid")
	}
	
	// Test transfer operations
	err = handle.Pause()
	if err != nil {
		t.Errorf("Pause() failed: %v", err)
	}
	
	status := handle.GetStatus()
	if status.State != TransferStatePaused {
		t.Errorf("Transfer state should be paused, got %v", status.State)
	}
	
	err = handle.Resume()
	if err != nil {
		t.Errorf("Resume() failed: %v", err)
	}
	
	// Test session stats
	stats := session.GetSessionStats()
	if stats.TotalTransfers != 1 {
		t.Errorf("Expected 1 transfer, got %d", stats.TotalTransfers)
	}
}

func TestPersistence(t *testing.T) {
	// Test memory persistence
	memPersistence := NewMemoryResumeData()
	
	link, _ := ParseEMuleLink("ed2k://|file|test.pdf|1048576|31D6CFE0D16AE931B73C59D7E0C089C0|/")
	resumeData := &TransferResumeData{
		Hash: link.Hash,
		Size: link.Size,
		Name: link.Name,
	}
	
	err := memPersistence.Save(link.Hash, resumeData)
	if err != nil {
		t.Errorf("Save() failed: %v", err)
	}
	
	if !memPersistence.Exists(link.Hash) {
		t.Error("Resume data should exist")
	}
	
	loaded, err := memPersistence.Load(link.Hash)
	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}
	
	if loaded.Hash.String() != link.Hash.String() {
		t.Error("Loaded hash doesn't match")
	}
	
	err = memPersistence.Remove(link.Hash)
	if err != nil {
		t.Errorf("Remove() failed: %v", err)
	}
	
	if memPersistence.Exists(link.Hash) {
		t.Error("Resume data should not exist after removal")
	}
}