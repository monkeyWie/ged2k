package jed2k

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
)

// LinkType represents the type of eMule link
type LinkType int

const (
	LinkTypeFile LinkType = iota
	LinkTypeServer
	LinkTypeSearch
)

// EMuleLink represents a parsed ed2k:// link
type EMuleLink struct {
	Hash        *hash.Hash `json:"hash"`
	Size        int64      `json:"size"`
	Name        string     `json:"name"`
	LinkType    LinkType   `json:"link_type"`
	Sources     []string   `json:"sources,omitempty"`
}

// ParseEMuleLink parses an ed2k:// URL into an EMuleLink
func ParseEMuleLink(linkStr string) (*EMuleLink, error) {
	if !strings.HasPrefix(linkStr, "ed2k://") {
		return nil, errors.New("invalid ed2k link: must start with ed2k://")
	}

	// Remove the protocol prefix
	content := strings.TrimPrefix(linkStr, "ed2k://")
	
	// ed2k links use pipe separators, so we need to parse them directly
	// Format: |type|param1|param2|...|/
	if !strings.HasPrefix(content, "|") || !strings.HasSuffix(content, "|/") {
		return nil, errors.New("invalid ed2k link format")
	}
	
	// Remove leading and trailing pipes
	content = strings.Trim(content, "|/")
	
	// Split by pipe
	segments := strings.Split(content, "|")
	if len(segments) < 1 {
		return nil, errors.New("invalid ed2k link: missing type")
	}

	switch segments[0] {
	case "file":
		return parseFileLink(segments, "")
	case "server":
		return parseServerLink(segments)
	case "search":
		return parseSearchLink(segments)
	default:
		return nil, fmt.Errorf("unsupported link type: %s", segments[0])
	}
}

// parseFileLink parses ed2k://|file|name|size|hash|sources|/
func parseFileLink(segments []string, query string) (*EMuleLink, error) {
	if len(segments) < 4 {
		return nil, errors.New("invalid file link: insufficient parameters")
	}

	emuleLink := &EMuleLink{
		LinkType: LinkTypeFile,
		Name:     segments[1],
	}

	// Parse size
	size, err := strconv.ParseInt(segments[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %v", err)
	}
	emuleLink.Size = size

	// Parse hash
	hashStr := segments[3]
	if len(hashStr) != 32 {
		return nil, errors.New("invalid hash: must be 32 characters")
	}
	
	hashBytes, err := hash.FromHexString(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %v", err)
	}
	emuleLink.Hash = hashBytes

	// Parse sources if present
	if len(segments) > 4 && segments[4] != "" {
		sources := strings.Split(segments[4], ",")
		emuleLink.Sources = sources
	}

	return emuleLink, nil
}

// parseServerLink parses ed2k://|server|ip|port|/
func parseServerLink(segments []string) (*EMuleLink, error) {
	if len(segments) < 3 {
		return nil, errors.New("invalid server link: insufficient parameters")
	}

	return &EMuleLink{
		LinkType: LinkTypeServer,
		Name:     segments[1], // IP
		Size:     0,           // Port will be in Size field for servers
	}, nil
}

// parseSearchLink parses ed2k://|search|query|/
func parseSearchLink(segments []string) (*EMuleLink, error) {
	if len(segments) < 2 {
		return nil, errors.New("invalid search link: missing query")
	}

	return &EMuleLink{
		LinkType: LinkTypeSearch,
		Name:     segments[1], // Search query
	}, nil
}

// String returns the string representation of the link
func (el *EMuleLink) String() string {
	switch el.LinkType {
	case LinkTypeFile:
		sourcesStr := ""
		if len(el.Sources) > 0 {
			sourcesStr = "|" + strings.Join(el.Sources, ",")
		}
		return fmt.Sprintf("ed2k://|file|%s|%d|%s%s|/", 
			el.Name, el.Size, el.Hash.String(), sourcesStr)
	case LinkTypeServer:
		return fmt.Sprintf("ed2k://|server|%s|%d|/", el.Name, el.Size)
	case LinkTypeSearch:
		return fmt.Sprintf("ed2k://|search|%s|/", el.Name)
	default:
		return ""
	}
}

// IsFileLink returns true if this is a file download link
func (el *EMuleLink) IsFileLink() bool {
	return el.LinkType == LinkTypeFile
}

// IsServerLink returns true if this is a server link
func (el *EMuleLink) IsServerLink() bool {
	return el.LinkType == LinkTypeServer
}

// IsSearchLink returns true if this is a search link
func (el *EMuleLink) IsSearchLink() bool {
	return el.LinkType == LinkTypeSearch
}