package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// KadContact represents a Kademlia contact
type KadContact struct {
	ID       [16]byte // 128-bit Kademlia ID
	IP       uint32
	TCPPort  uint16
	UDPPort  uint16
	Type     byte
	Version  byte
}

// NodesDat represents a nodes.dat file
type NodesDat struct {
	NumContacts      uint32
	Version          uint32
	BootstrapEdition uint32
	BootstrapEntries []KadContact
	Contacts         []KadContact
	ExtContacts      []KadExtContact
}

// KadExtContact represents extended contact information
type KadExtContact struct {
	Version byte
	UDPKey  uint32
}

// ParseNodesDat parses a nodes.dat file from raw data
func ParseNodesDat(data []byte) (*NodesDat, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid nodes.dat file: too short")
	}

	nodesDat := &NodesDat{}
	offset := 0

	// Parse number of contacts
	nodesDat.NumContacts = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Handle different file formats based on numContacts
	if nodesDat.NumContacts == 0 {
		// Version information present
		if len(data) < offset+4 {
			return nil, errors.New("insufficient data for version")
		}
		nodesDat.Version = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		if nodesDat.Version == 3 {
			// Bootstrap edition format
			if len(data) < offset+4 {
				return nil, errors.New("insufficient data for bootstrap edition")
			}
			nodesDat.BootstrapEdition = binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4

			if nodesDat.BootstrapEdition == 1 {
				// Parse bootstrap entries
				if len(data) < offset+4 {
					return nil, errors.New("insufficient data for bootstrap entry count")
				}
				numBootstrap := binary.LittleEndian.Uint32(data[offset : offset+4])
				offset += 4

				for i := uint32(0); i < numBootstrap; i++ {
					contact, newOffset, err := parseKadContact(data, offset)
					if err != nil {
						return nil, fmt.Errorf("error parsing bootstrap entry %d: %v", i, err)
					}
					nodesDat.BootstrapEntries = append(nodesDat.BootstrapEntries, contact)
					offset = newOffset
				}
				return nodesDat, nil
			}
		}

		// Re-read actual contact count for versions 1-3
		if nodesDat.Version >= 1 && nodesDat.Version <= 3 {
			if len(data) < offset+4 {
				return nil, errors.New("insufficient data for actual contact count")
			}
			nodesDat.NumContacts = binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4
		}
	}

	// Parse regular contacts
	for i := uint32(0); i < nodesDat.NumContacts; i++ {
		contact, newOffset, err := parseKadContact(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing contact %d: %v", i, err)
		}
		nodesDat.Contacts = append(nodesDat.Contacts, contact)
		offset = newOffset

		// Parse extended contact information for version >= 2
		if nodesDat.Version >= 2 {
			extContact, newOffset, err := parseKadExtContact(data, offset)
			if err != nil {
				return nil, fmt.Errorf("error parsing ext contact %d: %v", i, err)
			}
			nodesDat.ExtContacts = append(nodesDat.ExtContacts, extContact)
			offset = newOffset
		}
	}

	return nodesDat, nil
}

// parseKadContact parses a single Kademlia contact
func parseKadContact(data []byte, offset int) (KadContact, int, error) {
	contact := KadContact{}

	// Need 16 bytes for ID + 4 bytes for IP + 2 bytes for TCP port + 2 bytes for UDP port + 1 byte for type + 1 byte for version
	if len(data) < offset+26 {
		return contact, offset, errors.New("insufficient data for kad contact")
	}

	// Parse 128-bit Kademlia ID
	copy(contact.ID[:], data[offset:offset+16])
	offset += 16

	// Parse IP address
	contact.IP = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Parse TCP port
	contact.TCPPort = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse UDP port
	contact.UDPPort = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Parse type
	contact.Type = data[offset]
	offset++

	// Parse version
	contact.Version = data[offset]
	offset++

	return contact, offset, nil
}

// parseKadExtContact parses extended contact information
func parseKadExtContact(data []byte, offset int) (KadExtContact, int, error) {
	extContact := KadExtContact{}

	if len(data) < offset+5 {
		return extContact, offset, errors.New("insufficient data for extended contact")
	}

	// Parse version
	extContact.Version = data[offset]
	offset++

	// Parse UDP key
	extContact.UDPKey = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	return extContact, offset, nil
}

// DownloadNodesDat downloads and parses a nodes.dat file from URL
func DownloadNodesDat(url string) (*NodesDat, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download nodes.dat: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download nodes.dat: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read nodes.dat data: %v", err)
	}

	return ParseNodesDat(data)
}