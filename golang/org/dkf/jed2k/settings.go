package jed2k

import (
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/protocol"
)

// Settings contains configuration for the ed2k session
type Settings struct {
	UserAgent    *hash.Hash `json:"user_agent"`
	ModName      string     `json:"mod_name"`
	ClientName   string     `json:"client_name"`
	ListenPort   int        `json:"listen_port"`
	UDPPort      int        `json:"udp_port"`
	Version      int        `json:"version"`
	ModMajor     int        `json:"mod_major"`
	ModMinor     int        `json:"mod_minor"`
	ModBuild     int        `json:"mod_build"`
	MaxFailCount int        `json:"max_fail_count"`
	MaxPeerListSize int    `json:"max_peer_list_size"`
	MinPeerReconnectTime int `json:"min_peer_reconnect_time"`
	PeerConnectionTimeout int `json:"peer_connection_timeout"`
	SessionConnectionsLimit int `json:"session_connections_limit"`
	BufferPoolSize int     `json:"buffer_pool_size"`
	MaxConnectionsPerSecond int `json:"max_connections_per_second"`
	CompressionVersion int `json:"compression_version"`
	ServerSearchTimeout int `json:"server_search_timeout"`
	ReconnectToServer bool `json:"reconnect_to_server"`
	ServerPingTimeout int64 `json:"server_ping_timeout"`
	IncomingDirectory string `json:"incoming_directory"`
	ResumeDataDirectory string `json:"resume_data_directory"`
}

// NewDefaultSettings creates settings with default values
func NewDefaultSettings() *Settings {
	emuleHash := hash.NewEmuleHash()
	return &Settings{
		UserAgent:    emuleHash,
		ModName:      "ged2k",
		ClientName:   "ged2k",
		ListenPort:   4661,
		UDPPort:      4662,
		Version:      0x3c,
		ModMajor:     0,
		ModMinor:     0,
		ModBuild:     0,
		MaxFailCount: 20,
		MaxPeerListSize: 100,
		MinPeerReconnectTime: 10,
		PeerConnectionTimeout: 5,
		SessionConnectionsLimit: 20,
		BufferPoolSize: 250,
		MaxConnectionsPerSecond: 10,
		CompressionVersion: 0,
		ServerSearchTimeout: 15,
		ReconnectToServer: false,
		ServerPingTimeout: 0,
		IncomingDirectory: "./downloads",
		ResumeDataDirectory: "./resume",
	}
}

// ServerEndpoint represents a server connection point
type ServerEndpoint struct {
	Endpoint *protocol.Endpoint `json:"endpoint"`
	Name     string             `json:"name"`
}

// ServerList manages the list of available servers
type ServerList struct {
	servers []ServerEndpoint
}

// NewServerList creates a new server list
func NewServerList() *ServerList {
	return &ServerList{
		servers: make([]ServerEndpoint, 0),
	}
}

// AddServer adds a server to the list
func (sl *ServerList) AddServer(endpoint *protocol.Endpoint, name string) {
	sl.servers = append(sl.servers, ServerEndpoint{
		Endpoint: endpoint,
		Name:     name,
	})
}

// GetServers returns all servers
func (sl *ServerList) GetServers() []ServerEndpoint {
	return sl.servers
}

// LoadFromMet loads servers from a server.met file URL
func (sl *ServerList) LoadFromMet(url string) error {
	// TODO: Implement server.met file parsing
	// For now, add some default servers as examples
	defaultServers := []ServerEndpoint{
		{
			Endpoint: protocol.NewEndpointFromIPPort(uint32(176)<<24|uint32(103)<<16|uint32(48)<<8|uint32(36), 4184),
			Name:     "DonkeyServer No1",
		},
		{
			Endpoint: protocol.NewEndpointFromIPPort(uint32(195)<<24|uint32(245)<<16|uint32(244)<<8|uint32(205), 4661),
			Name:     "eMule Security No1",
		},
	}
	
	sl.servers = append(sl.servers, defaultServers...)
	return nil
}

// NodesData manages Kademlia nodes information
type NodesData struct {
	nodes []protocol.Endpoint
}

// NewNodesData creates a new nodes data container
func NewNodesData() *NodesData {
	return &NodesData{
		nodes: make([]protocol.Endpoint, 0),
	}
}

// LoadFromDat loads nodes from a nodes.dat file URL
func (nd *NodesData) LoadFromDat(url string) error {
	// TODO: Implement nodes.dat file parsing
	// For now, add some default nodes as examples
	defaultNodes := []protocol.Endpoint{
		*protocol.NewEndpointFromIPPort(uint32(195)<<24|uint32(245)<<16|uint32(244)<<8|uint32(205), 4665),
		*protocol.NewEndpointFromIPPort(uint32(176)<<24|uint32(103)<<16|uint32(48)<<8|uint32(36), 4665),
	}
	
	nd.nodes = append(nd.nodes, defaultNodes...)
	return nil
}

// GetNodes returns all nodes
func (nd *NodesData) GetNodes() []protocol.Endpoint {
	return nd.nodes
}