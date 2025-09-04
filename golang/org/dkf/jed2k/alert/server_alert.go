package alert

import "fmt"

// ServerAlert represents an alert related to server operations
type ServerAlert struct {
	*BaseAlert
	Identifier string
}

// NewServerAlert creates a new server alert
func NewServerAlert(id string) *ServerAlert {
	if id == "" {
		panic("identifier cannot be empty")
	}
	return &ServerAlert{
		BaseAlert:  NewBaseAlert(),
		Identifier: id,
	}
}

// Severity returns the alert severity
func (sa *ServerAlert) Severity() Severity {
	return Info
}

// Category returns the alert category
func (sa *ServerAlert) Category() int {
	return int(ServerNotification)
}

// String returns string representation
func (sa *ServerAlert) String() string {
	return fmt.Sprintf("ServerAlert{identifier=%s}", sa.Identifier)
}

// ServerConnectionAlert represents server connection alert
type ServerConnectionAlert struct {
	*ServerAlert
	Connected bool
}

// NewServerConnectionAlert creates a new server connection alert
func NewServerConnectionAlert(id string, connected bool) *ServerConnectionAlert {
	return &ServerConnectionAlert{
		ServerAlert: NewServerAlert(id),
		Connected:   connected,
	}
}

// String returns string representation
func (sca *ServerConnectionAlert) String() string {
	status := "disconnected"
	if sca.Connected {
		status = "connected"
	}
	return fmt.Sprintf("ServerConnectionAlert{identifier=%s, status=%s}", sca.Identifier, status)
}

// ServerMessageAlert represents server message alert
type ServerMessageAlert struct {
	*ServerAlert
	Message string
}

// NewServerMessageAlert creates a new server message alert
func NewServerMessageAlert(id string, message string) *ServerMessageAlert {
	return &ServerMessageAlert{
		ServerAlert: NewServerAlert(id),
		Message:     message,
	}
}

// String returns string representation
func (sma *ServerMessageAlert) String() string {
	return fmt.Sprintf("ServerMessageAlert{identifier=%s, message=%s}", sma.Identifier, sma.Message)
}

// ServerStatusAlert represents server status alert
type ServerStatusAlert struct {
	*ServerAlert
	Status string
}

// NewServerStatusAlert creates a new server status alert
func NewServerStatusAlert(id string, status string) *ServerStatusAlert {
	return &ServerStatusAlert{
		ServerAlert: NewServerAlert(id),
		Status:      status,
	}
}

// String returns string representation
func (ssa *ServerStatusAlert) String() string {
	return fmt.Sprintf("ServerStatusAlert{identifier=%s, status=%s}", ssa.Identifier, ssa.Status)
}