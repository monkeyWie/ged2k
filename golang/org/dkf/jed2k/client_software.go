package jed2k

// ClientSoftware represents different ed2k client software types
type ClientSoftware int

const (
	SOEmule         ClientSoftware = 0
	SOCDonkey       ClientSoftware = 1
	SOLXMule        ClientSoftware = 2
	SOAMule         ClientSoftware = 3
	SOShareaza      ClientSoftware = 4
	SOEmulePlus     ClientSoftware = 5
	SOHydranode     ClientSoftware = 6
	SONew2MLDonkey  ClientSoftware = 0x0a
	SOLPhant        ClientSoftware = 0x14
	SONew2Shareaza  ClientSoftware = 0x28
	SOEDonkeyHybrid ClientSoftware = 0x32
	SOEDonkey       ClientSoftware = 0x33
	SOMLDonkey      ClientSoftware = 0x34
	SOOldEmule      ClientSoftware = 0x35
	SOUnknown       ClientSoftware = 0x36
	SONewShareaza   ClientSoftware = 0x44
	SONewMLDonkey   ClientSoftware = 0x98
	SOLibED2K       ClientSoftware = 0x99
	SOQMule         ClientSoftware = 0xA0
	SOCompatUnk     ClientSoftware = 0xFF
)

// String returns string representation of client software
func (cs ClientSoftware) String() string {
	switch cs {
	case SOEmule:
		return "eMule"
	case SOCDonkey:
		return "cDonkey"
	case SOLXMule:
		return "lxMule"
	case SOAMule:
		return "aMule"
	case SOShareaza:
		return "Shareaza"
	case SOEmulePlus:
		return "eMule Plus"
	case SOHydranode:
		return "Hydranode"
	case SONew2MLDonkey:
		return "MLDonkey (new2)"
	case SOLPhant:
		return "lPhant"
	case SONew2Shareaza:
		return "Shareaza (new2)"
	case SOEDonkeyHybrid:
		return "eDonkey Hybrid"
	case SOEDonkey:
		return "eDonkey"
	case SOMLDonkey:
		return "MLDonkey"
	case SOOldEmule:
		return "eMule (old)"
	case SOUnknown:
		return "Unknown"
	case SONewShareaza:
		return "Shareaza (new)"
	case SONewMLDonkey:
		return "MLDonkey (new)"
	case SOLibED2K:
		return "libED2K"
	case SOQMule:
		return "qMule"
	case SOCompatUnk:
		return "Compatible Unknown"
	default:
		return "Undefined"
	}
}

// Value returns the numeric value
func (cs ClientSoftware) Value() int {
	return int(cs)
}