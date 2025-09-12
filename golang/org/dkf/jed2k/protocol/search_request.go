package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// FileType represents ed2k file type values
type FileType uint8

const (
	ED2KFTAny             FileType = 0
	ED2KFTAudio          FileType = 1
	ED2KFTVideo          FileType = 2
	ED2KFTImage          FileType = 3
	ED2KFTProgram        FileType = 4
	ED2KFTDocument       FileType = 5
	ED2KFTArchive        FileType = 6
	ED2KFTCdimage        FileType = 7
	ED2KFTEmuleCollection FileType = 8
)

// SearchOperator represents search comparison operators
type SearchOperator uint8

const (
	ED2KSearchOpEqual        SearchOperator = 0
	ED2KSearchOpGreater      SearchOperator = 1
	ED2KSearchOpLess         SearchOperator = 2
	ED2KSearchOpGreaterEqual SearchOperator = 3
	ED2KSearchOpLessEqual    SearchOperator = 4
	ED2KSearchOpNotEqual     SearchOperator = 5
)

// BooleanOperator represents boolean operators in search expressions
type BooleanOperator uint8

const (
	OperAnd BooleanOperator = 0
	OperOr  BooleanOperator = 1
	OperNot BooleanOperator = 2
)

// Search entry types
const (
	SearchTypeBool     uint8 = 0x00
	SearchTypeStr      uint8 = 0x01
	SearchTypeStrTag   uint8 = 0x02
	SearchTypeUInt32   uint8 = 0x03
	SearchTypeUInt64   uint8 = 0x08
)

// File type strings for searches
const (
	ED2KFTStrAudio            = "Audio"
	ED2KFTStrVideo            = "Video" 
	ED2KFTStrImage            = "Image"
	ED2KFTStrDocument         = "Doc"
	ED2KFTStrProgram          = "Pro"
	ED2KFTStrArchive          = "Arc"
	ED2KFTStrCdimage          = "Iso"
	ED2KFTStrEmuleCollection  = "EmuleCollection"
	ED2KFTStrFolder           = "Folder"
	ED2KFTStrUser             = "User"
)

// Tag IDs for file properties
const (
	FTFilename       uint8 = 0x01
	FTFilesize       uint8 = 0x02
	FTFiletype       uint8 = 0x03
	FTFileformat     uint8 = 0x04
	FTSources        uint8 = 0x15
	FTCompleteSources uint8 = 0x30
	FTMediaLength    uint8 = 0xD3
	FTMediaBitrate   uint8 = 0xD4
	FTMediaCodec     uint8 = 0xD5
)

// SearchExpression represents a single search expression entry
type SearchExpression interface {
	Serializable
	String() string
}

// StringEntry represents a string search term
type StringEntry struct {
	Value string
	Tag   *TagDescriptor
}

func (se *StringEntry) Put(dst *bytes.Buffer) error {
	// Write string length and value
	valueBytes := []byte(se.Value)
	if err := binary.Write(dst, binary.LittleEndian, uint16(len(valueBytes))); err != nil {
		return err
	}
	if _, err := dst.Write(valueBytes); err != nil {
		return err
	}
	
	// Write tag if present
	if se.Tag != nil {
		if err := se.Tag.Put(dst); err != nil {
			return err
		}
	} else {
		// Write empty tag
		if err := binary.Write(dst, binary.LittleEndian, uint16(0)); err != nil {
			return err
		}
	}
	return nil
}

func (se *StringEntry) Get(src *bytes.Buffer) error {
	var length uint16
	if err := binary.Read(src, binary.LittleEndian, &length); err != nil {
		return err
	}
	
	valueBytes := make([]byte, length)
	if _, err := io.ReadFull(src, valueBytes); err != nil {
		return err
	}
	se.Value = string(valueBytes)
	
	// Read tag
	tag := &TagDescriptor{}
	if err := tag.Get(src); err != nil {
		return err
	}
	if tag.HasData() {
		se.Tag = tag
	}
	
	return nil
}

func (se *StringEntry) BytesCount() int {
	size := 2 + len(se.Value) // string length + value
	if se.Tag != nil {
		size += se.Tag.BytesCount()
	} else {
		size += 2 // empty tag
	}
	return size
}

func (se *StringEntry) String() string {
	if se.Tag != nil {
		return fmt.Sprintf("StringEntry{%s, tag=%s}", se.Value, se.Tag)
	}
	return fmt.Sprintf("StringEntry{%s}", se.Value)
}

// NumericEntry represents a numeric comparison search term
type NumericEntry struct {
	Value    uint64
	Operator SearchOperator
	Tag      *TagDescriptor
}

func (ne *NumericEntry) Put(dst *bytes.Buffer) error {
	// Write operator
	if err := binary.Write(dst, binary.LittleEndian, uint8(ne.Operator)); err != nil {
		return err
	}
	
	// Write value as uint64
	if err := binary.Write(dst, binary.LittleEndian, ne.Value); err != nil {
		return err
	}
	
	// Write tag
	if ne.Tag != nil {
		if err := ne.Tag.Put(dst); err != nil {
			return err
		}
	} else {
		if err := binary.Write(dst, binary.LittleEndian, uint16(0)); err != nil {
			return err
		}
	}
	
	return nil
}

func (ne *NumericEntry) Get(src *bytes.Buffer) error {
	var op uint8
	if err := binary.Read(src, binary.LittleEndian, &op); err != nil {
		return err
	}
	ne.Operator = SearchOperator(op)
	
	if err := binary.Read(src, binary.LittleEndian, &ne.Value); err != nil {
		return err
	}
	
	tag := &TagDescriptor{}
	if err := tag.Get(src); err != nil {
		return err
	}
	if tag.HasData() {
		ne.Tag = tag
	}
	
	return nil
}

func (ne *NumericEntry) BytesCount() int {
	size := 1 + 8 // operator + uint64
	if ne.Tag != nil {
		size += ne.Tag.BytesCount()
	} else {
		size += 2
	}
	return size
}

func (ne *NumericEntry) String() string {
	opStr := map[SearchOperator]string{
		ED2KSearchOpEqual:        "=",
		ED2KSearchOpGreater:      ">",
		ED2KSearchOpLess:         "<",
		ED2KSearchOpGreaterEqual: ">=",
		ED2KSearchOpLessEqual:    "<=",
		ED2KSearchOpNotEqual:     "!=",
	}[ne.Operator]
	
	if ne.Tag != nil {
		return fmt.Sprintf("NumericEntry{%d %s, tag=%s}", ne.Value, opStr, ne.Tag)
	}
	return fmt.Sprintf("NumericEntry{%d %s}", ne.Value, opStr)
}

// BooleanEntry represents a boolean operator
type BooleanEntry struct {
	Operator BooleanOperator
}

func (be *BooleanEntry) Put(dst *bytes.Buffer) error {
	return binary.Write(dst, binary.LittleEndian, uint8(be.Operator))
}

func (be *BooleanEntry) Get(src *bytes.Buffer) error {
	var op uint8
	if err := binary.Read(src, binary.LittleEndian, &op); err != nil {
		return err
	}
	be.Operator = BooleanOperator(op)
	return nil
}

func (be *BooleanEntry) BytesCount() int {
	return 1
}

func (be *BooleanEntry) String() string {
	opStr := map[BooleanOperator]string{
		OperAnd: "AND",
		OperOr:  "OR", 
		OperNot: "NOT",
	}[be.Operator]
	return fmt.Sprintf("BooleanEntry{%s}", opStr)
}

// OpenParen represents an opening parenthesis
type OpenParen struct{}

func (op *OpenParen) Put(dst *bytes.Buffer) error {
	// Open paren is implicit in the expression structure
	return nil
}

func (op *OpenParen) Get(src *bytes.Buffer) error {
	return nil
}

func (op *OpenParen) BytesCount() int {
	return 0
}

func (op *OpenParen) String() string {
	return "OpenParen{(}"
}

// CloseParen represents a closing parenthesis
type CloseParen struct{}

func (cp *CloseParen) Put(dst *bytes.Buffer) error {
	// Close paren is implicit in the expression structure
	return nil
}

func (cp *CloseParen) Get(src *bytes.Buffer) error {
	return nil
}

func (cp *CloseParen) BytesCount() int {
	return 0
}

func (cp *CloseParen) String() string {
	return "CloseParen{)}"
}

// TagDescriptor represents a search tag
type TagDescriptor struct {
	Name   string
	ID     uint8
	HasID  bool
}

func (td *TagDescriptor) Put(dst *bytes.Buffer) error {
	if td.HasID {
		// Write tag with ID
		if err := binary.Write(dst, binary.LittleEndian, uint16(1)); err != nil {
			return err
		}
		if err := binary.Write(dst, binary.LittleEndian, td.ID); err != nil {
			return err
		}
	} else if td.Name != "" {
		// Write tag with name
		nameBytes := []byte(td.Name)
		if err := binary.Write(dst, binary.LittleEndian, uint16(len(nameBytes))); err != nil {
			return err
		}
		if _, err := dst.Write(nameBytes); err != nil {
			return err
		}
	} else {
		// Empty tag
		if err := binary.Write(dst, binary.LittleEndian, uint16(0)); err != nil {
			return err
		}
	}
	return nil
}

func (td *TagDescriptor) Get(src *bytes.Buffer) error {
	var length uint16
	if err := binary.Read(src, binary.LittleEndian, &length); err != nil {
		return err
	}
	
	if length == 0 {
		// Empty tag
		return nil
	} else if length == 1 {
		// Tag ID
		if err := binary.Read(src, binary.LittleEndian, &td.ID); err != nil {
			return err
		}
		td.HasID = true
	} else {
		// Tag name
		nameBytes := make([]byte, length)
		if _, err := io.ReadFull(src, nameBytes); err != nil {
			return err
		}
		td.Name = string(nameBytes)
		td.HasID = false
	}
	
	return nil
}

func (td *TagDescriptor) BytesCount() int {
	if td.HasID {
		return 2 + 1 // length + ID
	} else if td.Name != "" {
		return 2 + len(td.Name) // length + name
	}
	return 2 // empty tag
}

func (td *TagDescriptor) HasData() bool {
	return td.HasID || td.Name != ""
}

func (td *TagDescriptor) String() string {
	if td.HasID {
		return fmt.Sprintf("Tag{ID=%d}", td.ID)
	} else if td.Name != "" {
		return fmt.Sprintf("Tag{Name=%s}", td.Name)
	}
	return "Tag{empty}"
}

// SearchRequest represents a complete search request
type SearchRequest struct {
	Entries []SearchExpression
}

func NewSearchRequest() *SearchRequest {
	return &SearchRequest{
		Entries: make([]SearchExpression, 0),
	}
}

func (sr *SearchRequest) Put(dst *bytes.Buffer) error {
	for _, entry := range sr.Entries {
		if err := entry.Put(dst); err != nil {
			return err
		}
	}
	return nil
}

func (sr *SearchRequest) Get(src *bytes.Buffer) error {
	// This would need more complex parsing logic based on the protocol
	// For now, implement basic functionality
	return fmt.Errorf("SearchRequest.Get not implemented")
}

func (sr *SearchRequest) BytesCount() int {
	total := 0
	for _, entry := range sr.Entries {
		total += entry.BytesCount()
	}
	return total
}

func (sr *SearchRequest) String() string {
	var parts []string
	for _, entry := range sr.Entries {
		parts = append(parts, entry.String())
	}
	return fmt.Sprintf("SearchRequest{%s}", strings.Join(parts, ", "))
}

// AddString adds a string search term
func (sr *SearchRequest) AddString(value string) {
	sr.Entries = append(sr.Entries, &StringEntry{Value: value})
}

// AddStringWithTag adds a string search term with a tag
func (sr *SearchRequest) AddStringWithTag(value string, tagName string, tagID uint8, hasID bool) {
	tag := &TagDescriptor{Name: tagName, ID: tagID, HasID: hasID}
	sr.Entries = append(sr.Entries, &StringEntry{Value: value, Tag: tag})
}

// AddNumeric adds a numeric comparison term
func (sr *SearchRequest) AddNumeric(value uint64, op SearchOperator, tagID uint8) {
	tag := &TagDescriptor{ID: tagID, HasID: true}
	sr.Entries = append(sr.Entries, &NumericEntry{Value: value, Operator: op, Tag: tag})
}

// AddBoolean adds a boolean operator
func (sr *SearchRequest) AddBoolean(op BooleanOperator) {
	sr.Entries = append(sr.Entries, &BooleanEntry{Operator: op})
}

// MakeSimpleSearchRequest creates a simple text-based search request
func MakeSimpleSearchRequest(query string, fileType string, minSize, maxSize uint64) *SearchRequest {
	sr := NewSearchRequest()
	
	// Add file type filter if specified
	if fileType != "" {
		sr.AddStringWithTag(fileType, "", FTFiletype, true)
	}
	
	// Add size filters
	if minSize > 0 {
		sr.AddNumeric(minSize, ED2KSearchOpGreaterEqual, FTFilesize)
	}
	
	if maxSize > 0 {
		sr.AddNumeric(maxSize, ED2KSearchOpLessEqual, FTFilesize)
	}
	
	// Add query terms
	queryTerms := strings.Fields(strings.ToLower(query))
	for i, term := range queryTerms {
		if i > 0 {
			sr.AddBoolean(OperAnd)
		}
		sr.AddString(term)
	}
	
	return sr
}

// Count returns the number of entries in the search request
func (sr *SearchRequest) Count() int {
	return len(sr.Entries)
}

// Entry returns the entry at the specified index
func (sr *SearchRequest) Entry(index int) SearchExpression {
	if index >= 0 && index < len(sr.Entries) {
		return sr.Entries[index]
	}
	return nil
}