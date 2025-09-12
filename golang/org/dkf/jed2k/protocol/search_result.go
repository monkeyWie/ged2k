package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// SearchEntry represents a single file entry in search results
type SearchEntry struct {
	Hash             *Hash     `json:"hash"`
	ClientID         uint32    `json:"client_id"`
	ClientPort       uint16    `json:"client_port"`
	Filename         string    `json:"filename"`
	Filesize         uint64    `json:"filesize"`
	Filetype         string    `json:"filetype,omitempty"`
	Sources          uint32    `json:"sources"`
	CompleteSources  uint32    `json:"complete_sources,omitempty"`
	MediaBitrate     uint32    `json:"media_bitrate,omitempty"`
	MediaLength      uint32    `json:"media_length,omitempty"`
	MediaCodec       string    `json:"media_codec,omitempty"`
	Tags             []Tag     `json:"tags,omitempty"`
}

// SearchResult represents a search result packet from server
type SearchResult struct {
	Results     []*SearchEntry `json:"results"`
	MoreResults bool          `json:"more_results"`
}

func NewSearchResult() *SearchResult {
	return &SearchResult{
		Results: make([]*SearchEntry, 0),
	}
}

func (sr *SearchResult) Put(dst *bytes.Buffer) error {
	// Write number of results
	if err := binary.Write(dst, binary.LittleEndian, uint32(len(sr.Results))); err != nil {
		return err
	}
	
	// Write each result
	for _, result := range sr.Results {
		if err := result.Put(dst); err != nil {
			return err
		}
	}
	
	// Write more results flag
	moreFlag := byte(0)
	if sr.MoreResults {
		moreFlag = 1
	}
	if err := binary.Write(dst, binary.LittleEndian, moreFlag); err != nil {
		return err
	}
	
	return nil
}

func (sr *SearchResult) Get(src *bytes.Buffer) error {
	// Read number of results
	var count uint32
	if err := binary.Read(src, binary.LittleEndian, &count); err != nil {
		return err
	}
	
	// Read each result
	sr.Results = make([]*SearchEntry, count)
	for i := uint32(0); i < count; i++ {
		entry := &SearchEntry{}
		if err := entry.Get(src); err != nil {
			return err
		}
		sr.Results[i] = entry
	}
	
	// Read more results flag
	var moreFlag byte
	if err := binary.Read(src, binary.LittleEndian, &moreFlag); err != nil {
		return err
	}
	sr.MoreResults = moreFlag != 0
	
	return nil
}

func (sr *SearchResult) BytesCount() int {
	size := 4 + 1 // count + more flag
	for _, result := range sr.Results {
		size += result.BytesCount()
	}
	return size
}

func (sr *SearchResult) String() string {
	return fmt.Sprintf("SearchResult{%d results, more=%v}", len(sr.Results), sr.MoreResults)
}

// GetResults returns the search results
func (sr *SearchResult) GetResults() []*SearchEntry {
	return sr.Results
}

// HasMoreResults returns true if more results are available
func (sr *SearchResult) HasMoreResults() bool {
	return sr.MoreResults
}

// AddResult adds a search result entry
func (sr *SearchResult) AddResult(entry *SearchEntry) {
	sr.Results = append(sr.Results, entry)
}

// SearchEntry implementation

func NewSearchEntry() *SearchEntry {
	return &SearchEntry{
		Tags: make([]Tag, 0),
	}
}

func (se *SearchEntry) Put(dst *bytes.Buffer) error {
	// Write hash
	if se.Hash == nil {
		return fmt.Errorf("search entry hash is nil")
	}
	if err := se.Hash.Put(dst); err != nil {
		return err
	}
	
	// Write client ID and port
	if err := binary.Write(dst, binary.LittleEndian, se.ClientID); err != nil {
		return err
	}
	if err := binary.Write(dst, binary.LittleEndian, se.ClientPort); err != nil {
		return err
	}
	
	// Write number of tags
	tagCount := 0
	if se.Filename != "" { tagCount++ }
	if se.Filesize > 0 { tagCount++ }
	if se.Filetype != "" { tagCount++ }
	if se.Sources > 0 { tagCount++ }
	if se.CompleteSources > 0 { tagCount++ }
	if se.MediaBitrate > 0 { tagCount++ }
	if se.MediaLength > 0 { tagCount++ }
	if se.MediaCodec != "" { tagCount++ }
	tagCount += len(se.Tags)
	
	if err := binary.Write(dst, binary.LittleEndian, uint32(tagCount)); err != nil {
		return err
	}
	
	// Write standard tags
	if se.Filename != "" {
		if err := writeStringTag(dst, FTFilename, se.Filename); err != nil {
			return err
		}
	}
	if se.Filesize > 0 {
		if err := writeUInt64Tag(dst, FTFilesize, se.Filesize); err != nil {
			return err
		}
	}
	if se.Filetype != "" {
		if err := writeStringTag(dst, FTFiletype, se.Filetype); err != nil {
			return err
		}
	}
	if se.Sources > 0 {
		if err := writeUInt32Tag(dst, FTSources, se.Sources); err != nil {
			return err
		}
	}
	if se.CompleteSources > 0 {
		if err := writeUInt32Tag(dst, FTCompleteSources, se.CompleteSources); err != nil {
			return err
		}
	}
	if se.MediaBitrate > 0 {
		if err := writeUInt32Tag(dst, FTMediaBitrate, se.MediaBitrate); err != nil {
			return err
		}
	}
	if se.MediaLength > 0 {
		if err := writeUInt32Tag(dst, FTMediaLength, se.MediaLength); err != nil {
			return err
		}
	}
	if se.MediaCodec != "" {
		if err := writeStringTag(dst, FTMediaCodec, se.MediaCodec); err != nil {
			return err
		}
	}
	
	// Write additional tags
	for _, tag := range se.Tags {
		if err := tag.Put(dst); err != nil {
			return err
		}
	}
	
	return nil
}

func (se *SearchEntry) Get(src *bytes.Buffer) error {
	// Read hash
	se.Hash = NewHash()
	if err := se.Hash.Get(src); err != nil {
		return err
	}
	
	// Read client ID and port
	if err := binary.Read(src, binary.LittleEndian, &se.ClientID); err != nil {
		return err
	}
	if err := binary.Read(src, binary.LittleEndian, &se.ClientPort); err != nil {
		return err
	}
	
	// Read number of tags
	var tagCount uint32
	if err := binary.Read(src, binary.LittleEndian, &tagCount); err != nil {
		return err
	}
	
	// Read tags
	for i := uint32(0); i < tagCount; i++ {
		tag := Tag{}
		if err := tag.Get(src); err != nil {
			return err
		}
		
		// Process standard tags
		switch tag.Type {
		case TagTypeString:
			switch tag.Name {
			case string([]byte{FTFilename}):
				if str, ok := tag.Value.(string); ok {
					se.Filename = str
				}
			case string([]byte{FTFiletype}):
				if str, ok := tag.Value.(string); ok {
					se.Filetype = str
				}
			case string([]byte{FTMediaCodec}):
				if str, ok := tag.Value.(string); ok {
					se.MediaCodec = str
				}
			default:
				se.Tags = append(se.Tags, tag)
			}
		case TagTypeUInt32:
			switch tag.Name {
			case string([]byte{FTSources}):
				if val, ok := tag.Value.(uint32); ok {
					se.Sources = val
				}
			case string([]byte{FTCompleteSources}):
				if val, ok := tag.Value.(uint32); ok {
					se.CompleteSources = val
				}
			case string([]byte{FTMediaBitrate}):
				if val, ok := tag.Value.(uint32); ok {
					se.MediaBitrate = val
				}
			case string([]byte{FTMediaLength}):
				if val, ok := tag.Value.(uint32); ok {
					se.MediaLength = val
				}
			default:
				se.Tags = append(se.Tags, tag)
			}
		case TagTypeUInt64:
			switch tag.Name {
			case string([]byte{FTFilesize}):
				if val, ok := tag.Value.(uint64); ok {
					se.Filesize = val
				}
			default:
				se.Tags = append(se.Tags, tag)
			}
		default:
			se.Tags = append(se.Tags, tag)
		}
	}
	
	return nil
}

func (se *SearchEntry) BytesCount() int {
	size := HashSize + 4 + 2 + 4 // hash + clientID + clientPort + tag count
	
	// Standard tags
	if se.Filename != "" { size += calculateStringTagSize(se.Filename) }
	if se.Filesize > 0 { size += calculateUInt64TagSize() }
	if se.Filetype != "" { size += calculateStringTagSize(se.Filetype) }
	if se.Sources > 0 { size += calculateUInt32TagSize() }
	if se.CompleteSources > 0 { size += calculateUInt32TagSize() }
	if se.MediaBitrate > 0 { size += calculateUInt32TagSize() }
	if se.MediaLength > 0 { size += calculateUInt32TagSize() }
	if se.MediaCodec != "" { size += calculateStringTagSize(se.MediaCodec) }
	
	// Additional tags
	for _, tag := range se.Tags {
		size += tag.BytesCount()
	}
	
	return size
}

func (se *SearchEntry) String() string {
	return fmt.Sprintf("SearchEntry{name=%s, size=%d, sources=%d, hash=%s}", 
		se.Filename, se.Filesize, se.Sources, se.Hash.String()[:8]+"...")
}

// GetHash returns the file hash
func (se *SearchEntry) GetHash() *Hash {
	return se.Hash
}

// GetFilename returns the filename
func (se *SearchEntry) GetFilename() string {
	return se.Filename
}

// GetFilesize returns the file size
func (se *SearchEntry) GetFilesize() uint64 {
	return se.Filesize
}

// GetSources returns the number of sources
func (se *SearchEntry) GetSources() uint32 {
	return se.Sources
}

// GetEndpoint returns the client endpoint (IP:port)
func (se *SearchEntry) GetEndpoint() *Endpoint {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, se.ClientID)
	addr := &net.TCPAddr{IP: ip, Port: int(se.ClientPort)}
	return NewEndpointFromTCPAddr(addr)
}

// ToED2KLink converts the search entry to an ed2k link
func (se *SearchEntry) ToED2KLink() string {
	return fmt.Sprintf("ed2k://|file|%s|%d|%s|/", 
		se.Filename, se.Filesize, se.Hash.String())
}

// Helper functions for writing tags

func writeStringTag(dst *bytes.Buffer, tagID uint8, value string) error {
	tag := Tag{
		Type:  TagTypeString,
		Name:  string([]byte{tagID}),
		Value: value,
	}
	return tag.Put(dst)
}

func writeUInt32Tag(dst *bytes.Buffer, tagID uint8, value uint32) error {
	tag := Tag{
		Type:  TagTypeUInt32,
		Name:  string([]byte{tagID}),
		Value: value,
	}
	return tag.Put(dst)
}

func writeUInt64Tag(dst *bytes.Buffer, tagID uint8, value uint64) error {
	tag := Tag{
		Type:  TagTypeUInt64,
		Name:  string([]byte{tagID}),
		Value: value,
	}
	return tag.Put(dst)
}

func calculateStringTagSize(value string) int {
	return 1 + 2 + 1 + 2 + len(value) // type + name_len + name + value_len + value
}

func calculateUInt32TagSize() int {
	return 1 + 2 + 1 + 4 // type + name_len + name + value
}

func calculateUInt64TagSize() int {
	return 1 + 2 + 1 + 8 // type + name_len + name + value
}

// Utility functions

// SortByRelevance sorts search entries by relevance (sources, then size)
func (sr *SearchResult) SortByRelevance() {
	// Simple bubble sort for now - could be improved
	results := sr.Results
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			// Sort by sources (descending), then by size (descending)
			if results[j].Sources < results[j+1].Sources || 
			   (results[j].Sources == results[j+1].Sources && 
			    results[j].Filesize < results[j+1].Filesize) {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// FilterByFileType filters search results by file type
func (sr *SearchResult) FilterByFileType(fileType string) *SearchResult {
	filtered := NewSearchResult()
	filtered.MoreResults = sr.MoreResults
	
	for _, entry := range sr.Results {
		if fileType == "" || entry.Filetype == fileType {
			filtered.Results = append(filtered.Results, entry)
		}
	}
	
	return filtered
}

// FilterBySize filters search results by size range
func (sr *SearchResult) FilterBySize(minSize, maxSize uint64) *SearchResult {
	filtered := NewSearchResult()
	filtered.MoreResults = sr.MoreResults
	
	for _, entry := range sr.Results {
		if (minSize == 0 || entry.Filesize >= minSize) && 
		   (maxSize == 0 || entry.Filesize <= maxSize) {
			filtered.Results = append(filtered.Results, entry)
		}
	}
	
	return filtered
}

// GetFormattedResults returns search results in a human-readable format
func (sr *SearchResult) GetFormattedResults() string {
	if len(sr.Results) == 0 {
		return "No search results found"
	}
	
	result := fmt.Sprintf("Found %d result(s):\n", len(sr.Results))
	for i, entry := range sr.Results {
		result += fmt.Sprintf("[%d] %s (%.2f MB, %d sources) - %s\n",
			i+1,
			entry.Filename,
			float64(entry.Filesize)/(1024*1024),
			entry.Sources,
			entry.Hash.String()[:16]+"...")
	}
	
	if sr.MoreResults {
		result += "\nMore results available on server."
	}
	
	return result
}

// CreateMockSearchResult creates a mock search result for testing
func CreateMockSearchResult(query string, count int) *SearchResult {
	result := NewSearchResult()
	
	for i := 0; i < count; i++ {
		hash, _ := HashFromString(fmt.Sprintf("%032d", i))
		entry := &SearchEntry{
			Hash:            hash,
			ClientID:        0xC0A80100 + uint32(i), // 192.168.1.x IPs
			ClientPort:      4661,
			Filename:        fmt.Sprintf("%s_file_%d.mp3", query, i+1),
			Filesize:        uint64(1024*1024*(i+1)), // 1MB, 2MB, etc.
			Filetype:        ED2KFTStrAudio,
			Sources:         uint32(10 + i),
			CompleteSources: uint32(5 + i),
			MediaBitrate:    192,
			MediaLength:     180 + uint32(i*30), // 3 min + 30s per result
		}
		result.AddResult(entry)
	}
	
	result.MoreResults = count >= 50 // Simulate more results for larger sets
	
	return result
}