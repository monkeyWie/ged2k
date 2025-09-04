package jed2k

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// ResumeData interface for transfer resume data persistence
type ResumeData interface {
	// Save saves the transfer resume data
	Save(transferHash *hash.Hash, data *TransferResumeData) error
	
	// Load loads the transfer resume data
	Load(transferHash *hash.Hash) (*TransferResumeData, error)
	
	// Remove removes the transfer resume data
	Remove(transferHash *hash.Hash) error
	
	// Exists checks if resume data exists for the transfer
	Exists(transferHash *hash.Hash) bool
}

// TransferResumeData contains data needed to resume a transfer
type TransferResumeData struct {
	Hash              *hash.Hash              `json:"hash"`
	Size              int64                   `json:"size"`
	Name              string                  `json:"name"`
	DownloadDirectory string                  `json:"download_directory"`
	Downloaded        int64                   `json:"downloaded"`
	Pieces            []bool                  `json:"pieces"`           // which pieces are completed
	DownloadedBlocks  []BlockInfo             `json:"downloaded_blocks"` // partially downloaded blocks
	Peers             []*protocol.Endpoint    `json:"peers"`            // known peers
	CreateTime        int64                   `json:"create_time"`
	LastSeen          int64                   `json:"last_seen"`
}

// BlockInfo represents information about a downloaded block
type BlockInfo struct {
	PieceIndex int   `json:"piece_index"`
	Offset     int64 `json:"offset"`
	Length     int64 `json:"length"`
}

// MemoryResumeData implements ResumeData interface using in-memory storage
type MemoryResumeData struct {
	data map[string]*TransferResumeData
}

// NewMemoryResumeData creates a new memory-based resume data storage
func NewMemoryResumeData() ResumeData {
	return &MemoryResumeData{
		data: make(map[string]*TransferResumeData),
	}
}

// Save saves the transfer resume data in memory
func (m *MemoryResumeData) Save(transferHash *hash.Hash, data *TransferResumeData) error {
	key := transferHash.String()
	m.data[key] = data
	return nil
}

// Load loads the transfer resume data from memory
func (m *MemoryResumeData) Load(transferHash *hash.Hash) (*TransferResumeData, error) {
	key := transferHash.String()
	if data, exists := m.data[key]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("resume data not found for hash %s", key)
}

// Remove removes the transfer resume data from memory
func (m *MemoryResumeData) Remove(transferHash *hash.Hash) error {
	key := transferHash.String()
	delete(m.data, key)
	return nil
}

// Exists checks if resume data exists in memory
func (m *MemoryResumeData) Exists(transferHash *hash.Hash) bool {
	key := transferHash.String()
	_, exists := m.data[key]
	return exists
}

// DiskResumeData implements ResumeData interface using disk storage
type DiskResumeData struct {
	basePath string
}

// NewDiskResumeData creates a new disk-based resume data storage
func NewDiskResumeData(basePath string) ResumeData {
	// Create directory if it doesn't exist
	os.MkdirAll(basePath, 0755)
	
	return &DiskResumeData{
		basePath: basePath,
	}
}

// getFilePath returns the file path for a transfer's resume data
func (d *DiskResumeData) getFilePath(transferHash *hash.Hash) string {
	filename := transferHash.String() + ".resume"
	return filepath.Join(d.basePath, filename)
}

// Save saves the transfer resume data to disk
func (d *DiskResumeData) Save(transferHash *hash.Hash, data *TransferResumeData) error {
	filepath := d.getFilePath(transferHash)
	
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create resume file: %v", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode resume data: %v", err)
	}
	
	return nil
}

// Load loads the transfer resume data from disk
func (d *DiskResumeData) Load(transferHash *hash.Hash) (*TransferResumeData, error) {
	filepath := d.getFilePath(transferHash)
	
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("resume data not found for hash %s", transferHash.String())
		}
		return nil, fmt.Errorf("failed to open resume file: %v", err)
	}
	defer file.Close()
	
	var data TransferResumeData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode resume data: %v", err)
	}
	
	return &data, nil
}

// Remove removes the transfer resume data from disk
func (d *DiskResumeData) Remove(transferHash *hash.Hash) error {
	filepath := d.getFilePath(transferHash)
	
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove resume file: %v", err)
	}
	
	return nil
}

// Exists checks if resume data exists on disk
func (d *DiskResumeData) Exists(transferHash *hash.Hash) bool {
	filepath := d.getFilePath(transferHash)
	_, err := os.Stat(filepath)
	return err == nil
}

// AddTransferParams contains parameters for adding a new transfer
type AddTransferParams struct {
	Hash              *hash.Hash `json:"hash"`
	Size              int64      `json:"size"`
	Name              string     `json:"name"`
	DownloadDirectory string     `json:"download_directory"`
	Paused            bool       `json:"paused"`
	ResumeData        *TransferResumeData `json:"resume_data,omitempty"`
}

// NewAddTransferParams creates transfer parameters from an eMule link
func NewAddTransferParams(link *EMuleLink, downloadDir string) (*AddTransferParams, error) {
	if !link.IsFileLink() {
		return nil, fmt.Errorf("can only create transfer params from file links")
	}
	
	return &AddTransferParams{
		Hash:              link.Hash,
		Size:              link.Size,
		Name:              link.Name,
		DownloadDirectory: downloadDir,
		Paused:            false,
	}, nil
}

// FileHandler interface abstracts file operations for different platforms
type FileHandler interface {
	// Open opens the file for reading/writing
	Open() error
	
	// Close closes the file
	Close() error
	
	// Read reads data from the file at the specified offset
	Read(offset int64, data []byte) (int, error)
	
	// Write writes data to the file at the specified offset
	Write(offset int64, data []byte) (int, error)
	
	// Size returns the current file size
	Size() (int64, error)
	
	// Sync synchronizes the file to disk
	Sync() error
	
	// Remove removes the file
	Remove() error
	
	// GetPath returns the file path
	GetPath() string
}

// DefaultFileHandler implements FileHandler for standard file operations
type DefaultFileHandler struct {
	filepath string
	file     *os.File
}

// NewDefaultFileHandler creates a new default file handler
func NewDefaultFileHandler(filepath string) FileHandler {
	return &DefaultFileHandler{
		filepath: filepath,
	}
}

// Open opens the file
func (fh *DefaultFileHandler) Open() error {
	if fh.file != nil {
		return nil // already open
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fh.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	// Open file for read/write, create if doesn't exist
	file, err := os.OpenFile(fh.filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	
	fh.file = file
	return nil
}

// Close closes the file
func (fh *DefaultFileHandler) Close() error {
	if fh.file == nil {
		return nil
	}
	
	err := fh.file.Close()
	fh.file = nil
	return err
}

// Read reads data from the file
func (fh *DefaultFileHandler) Read(offset int64, data []byte) (int, error) {
	if fh.file == nil {
		return 0, fmt.Errorf("file not open")
	}
	
	return fh.file.ReadAt(data, offset)
}

// Write writes data to the file
func (fh *DefaultFileHandler) Write(offset int64, data []byte) (int, error) {
	if fh.file == nil {
		return 0, fmt.Errorf("file not open")
	}
	
	return fh.file.WriteAt(data, offset)
}

// Size returns the file size
func (fh *DefaultFileHandler) Size() (int64, error) {
	if fh.file == nil {
		return 0, fmt.Errorf("file not open")
	}
	
	stat, err := fh.file.Stat()
	if err != nil {
		return 0, err
	}
	
	return stat.Size(), nil
}

// Sync synchronizes the file to disk
func (fh *DefaultFileHandler) Sync() error {
	if fh.file == nil {
		return fmt.Errorf("file not open")
	}
	
	return fh.file.Sync()
}

// Remove removes the file
func (fh *DefaultFileHandler) Remove() error {
	fh.Close()
	return os.Remove(fh.filepath)
}

// GetPath returns the file path
func (fh *DefaultFileHandler) GetPath() string {
	return fh.filepath
}

// PersistenceManager manages all persistence operations
type PersistenceManager struct {
	resumeData ResumeData
}

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager(resumeData ResumeData) *PersistenceManager {
	return &PersistenceManager{
		resumeData: resumeData,
	}
}

// SaveTransfer saves transfer resume data
func (pm *PersistenceManager) SaveTransfer(transfer *Transfer) error {
	status := transfer.GetStatus()
	
	resumeData := &TransferResumeData{
		Hash:              status.Hash,
		Size:              status.Size,
		Name:              status.Name,
		DownloadDirectory: status.DownloadDirectory,
		Downloaded:        status.Downloaded,
		// TODO: Add pieces and blocks information
		Pieces:           make([]bool, 0),
		DownloadedBlocks: make([]BlockInfo, 0),
		Peers:            make([]*protocol.Endpoint, 0),
		CreateTime:       transfer.createTime.Unix(),
		LastSeen:         transfer.createTime.Unix(),
	}
	
	return pm.resumeData.Save(status.Hash, resumeData)
}

// LoadTransfer loads transfer resume data
func (pm *PersistenceManager) LoadTransfer(transferHash *hash.Hash) (*TransferResumeData, error) {
	return pm.resumeData.Load(transferHash)
}

// RemoveTransfer removes transfer resume data
func (pm *PersistenceManager) RemoveTransfer(transferHash *hash.Hash) error {
	return pm.resumeData.Remove(transferHash)
}

// HasTransfer checks if transfer resume data exists
func (pm *PersistenceManager) HasTransfer(transferHash *hash.Hash) bool {
	return pm.resumeData.Exists(transferHash)
}