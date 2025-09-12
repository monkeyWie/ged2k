package alert

import (
	"testing"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
)

func TestBaseAlert(t *testing.T) {
	alert := NewBaseAlert()
	
	if alert.GetCreationTime() <= 0 {
		t.Error("Creation time should be positive")
	}
}

func TestTransferAlerts(t *testing.T) {
	h := hash.NewHash()
	
	// Test TransferAlert
	alert := NewTransferAlert(h)
	if alert.Severity() != Info {
		t.Error("Transfer alert severity should be Info")
	}
	
	if alert.Category() != int(StatusNotification) {
		t.Error("Transfer alert category should be StatusNotification")
	}
	
	if alert.Hash != h {
		t.Error("Transfer alert hash should match")
	}
	
	// Test TransferAddedAlert
	addedAlert := NewTransferAddedAlert(h)
	if addedAlert.Severity() != Info {
		t.Error("Transfer added alert severity should be Info")
	}
	
	str := addedAlert.String()
	if len(str) == 0 {
		t.Error("Transfer added alert should have string representation")
	}
	
	// Test TransferFinishedAlert
	finishedAlert := NewTransferFinishedAlert(h)
	if finishedAlert.Severity() != Info {
		t.Error("Transfer finished alert severity should be Info")
	}
	
	// Test TransferPausedAlert
	pausedAlert := NewTransferPausedAlert(h)
	if pausedAlert.Severity() != Info {
		t.Error("Transfer paused alert severity should be Info")
	}
	
	// Test TransferResumedAlert
	resumedAlert := NewTransferResumedAlert(h)
	if resumedAlert.Severity() != Info {
		t.Error("Transfer resumed alert severity should be Info")
	}
	
	// Test TransferRemovedAlert
	removedAlert := NewTransferRemovedAlert(h)
	if removedAlert.Severity() != Info {
		t.Error("Transfer removed alert severity should be Info")
	}
}

func TestServerAlerts(t *testing.T) {
	id := "server1"
	
	// Test ServerAlert
	alert := NewServerAlert(id)
	if alert.Identifier != id {
		t.Error("Server alert identifier should match")
	}
	
	if alert.Severity() != Info {
		t.Error("Server alert severity should be Info")
	}
	
	if alert.Category() != int(ServerNotification) {
		t.Error("Server alert category should be ServerNotification")
	}
	
	// Test ServerConnectionAlert
	connAlert := NewServerConnectionAlert(id, true)
	if !connAlert.Connected {
		t.Error("Server connection alert should show connected")
	}
	
	str := connAlert.String()
	if len(str) == 0 {
		t.Error("Server connection alert should have string representation")
	}
	
	// Test ServerMessageAlert
	message := "test message"
	msgAlert := NewServerMessageAlert(id, message)
	if msgAlert.Message != message {
		t.Error("Server message alert message should match")
	}
	
	// Test ServerStatusAlert
	status := "online"
	statusAlert := NewServerStatusAlert(id, status)
	if statusAlert.Status != status {
		t.Error("Server status alert status should match")
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{Debug, "Debug"},
		{Info, "Info"},
		{Warning, "Warning"},
		{Critical, "Critical"},
		{Fatal, "Fatal"},
		{None, "None"},
	}
	
	for _, test := range tests {
		if test.severity.String() != test.expected {
			t.Errorf("Severity %d should be %s, got %s", test.severity, test.expected, test.severity.String())
		}
	}
}

func TestAlertPanic(t *testing.T) {
	// Test that empty identifier panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewServerAlert with empty identifier should panic")
		}
	}()
	
	NewServerAlert("")
}