package alert

import "github.com/monkeyWie/ged2k/golang/org/dkf/jed2k"

// Severity represents alert severity levels
type Severity int

const (
	Debug Severity = iota
	Info
	Warning
	Critical
	Fatal
	None
)

// String returns string representation of severity
func (s Severity) String() string {
	switch s {
	case Debug:
		return "Debug"
	case Info:
		return "Info"
	case Warning:
		return "Warning"
	case Critical:
		return "Critical"
	case Fatal:
		return "Fatal"
	case None:
		return "None"
	default:
		return "Unknown"
	}
}

// Category represents alert categories
type Category int

const (
	ErrorNotification       Category = 0x1
	PeerNotification        Category = 0x2
	PortMappingNotification Category = 0x4
	StorageNotification     Category = 0x8
	TrackerNotification     Category = 0x10
	DebugNotification       Category = 0x20
	StatusNotification      Category = 0x40
	ProgressNotification    Category = 0x80
	IpBlockNotification     Category = 0x100
	PerformanceWarning      Category = 0x200
	ServerNotification      Category = 0x400
	StatsNotification       Category = 0x800
	AllCategories           Category = 0xffffffff
)

// Alert represents the base alert interface
type Alert interface {
	Severity() Severity
	Category() int
	GetCreationTime() int64
	String() string
}

// BaseAlert provides common alert functionality
type BaseAlert struct {
	creationTime int64
}

// NewBaseAlert creates a new base alert
func NewBaseAlert() *BaseAlert {
	return &BaseAlert{
		creationTime: jed2k.CurrentTime(),
	}
}

// GetCreationTime returns the alert creation time
func (ba *BaseAlert) GetCreationTime() int64 {
	return ba.creationTime
}