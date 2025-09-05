package alert

import (
	"sync"
	"time"
)

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
		creationTime: time.Now().Unix(),
	}
}

// GetCreationTime returns the alert creation time
func (ba *BaseAlert) GetCreationTime() int64 {
	return ba.creationTime
}

// AlertManager manages alert posting and retrieval
type AlertManager struct {
	alerts  []Alert
	mutex   sync.RWMutex
	maxSize int
}

// NewAlertManager creates a new alert manager
func NewAlertManager(maxSize int) *AlertManager {
	if maxSize <= 0 {
		maxSize = 1000 // Default max size
	}
	return &AlertManager{
		alerts:  make([]Alert, 0, maxSize),
		maxSize: maxSize,
	}
}

// PostAlert adds an alert to the queue
func (am *AlertManager) PostAlert(alert Alert) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.alerts = append(am.alerts, alert)
	
	// Keep only the most recent alerts if we exceed max size
	if len(am.alerts) > am.maxSize {
		am.alerts = am.alerts[len(am.alerts)-am.maxSize:]
	}
}

// PopAlert retrieves and removes the oldest alert
func (am *AlertManager) PopAlert() Alert {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if len(am.alerts) == 0 {
		return nil
	}
	
	alert := am.alerts[0]
	am.alerts = am.alerts[1:]
	return alert
}

// GetAlerts returns all alerts without removing them
func (am *AlertManager) GetAlerts() []Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make([]Alert, len(am.alerts))
	copy(result, am.alerts)
	return result
}

// Clear removes all alerts
func (am *AlertManager) Clear() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.alerts = am.alerts[:0]
}

// Count returns the number of pending alerts
func (am *AlertManager) Count() int {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	return len(am.alerts)
}