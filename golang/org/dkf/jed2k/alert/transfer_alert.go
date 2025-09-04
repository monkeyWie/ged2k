package alert

import (
	"fmt"

	"github.com/monkeyWie/ged2k/golang/org/dkf/jed2k/hash"
)

// TransferAlert represents an alert related to a transfer
type TransferAlert struct {
	*BaseAlert
	Hash *hash.Hash
}

// NewTransferAlert creates a new transfer alert
func NewTransferAlert(h *hash.Hash) *TransferAlert {
	return &TransferAlert{
		BaseAlert: NewBaseAlert(),
		Hash:      h,
	}
}

// Severity returns the alert severity
func (ta *TransferAlert) Severity() Severity {
	return Info
}

// Category returns the alert category
func (ta *TransferAlert) Category() int {
	return int(StatusNotification)
}

// String returns string representation
func (ta *TransferAlert) String() string {
	if ta.Hash != nil {
		return ta.Hash.String()
	}
	return "TransferAlert{hash=nil}"
}

// TransferAddedAlert represents a transfer added alert
type TransferAddedAlert struct {
	*TransferAlert
}

// NewTransferAddedAlert creates a new transfer added alert
func NewTransferAddedAlert(h *hash.Hash) *TransferAddedAlert {
	return &TransferAddedAlert{
		TransferAlert: NewTransferAlert(h),
	}
}

// Severity returns the alert severity
func (taa *TransferAddedAlert) Severity() Severity {
	return Info
}

// String returns string representation
func (taa *TransferAddedAlert) String() string {
	return fmt.Sprintf("TransferAddedAlert{hash=%s}", taa.Hash.String())
}

// TransferFinishedAlert represents a transfer finished alert
type TransferFinishedAlert struct {
	*TransferAlert
}

// NewTransferFinishedAlert creates a new transfer finished alert
func NewTransferFinishedAlert(h *hash.Hash) *TransferFinishedAlert {
	return &TransferFinishedAlert{
		TransferAlert: NewTransferAlert(h),
	}
}

// Severity returns the alert severity
func (tfa *TransferFinishedAlert) Severity() Severity {
	return Info
}

// String returns string representation
func (tfa *TransferFinishedAlert) String() string {
	return fmt.Sprintf("TransferFinishedAlert{hash=%s}", tfa.Hash.String())
}

// TransferPausedAlert represents a transfer paused alert
type TransferPausedAlert struct {
	*TransferAlert
}

// NewTransferPausedAlert creates a new transfer paused alert
func NewTransferPausedAlert(h *hash.Hash) *TransferPausedAlert {
	return &TransferPausedAlert{
		TransferAlert: NewTransferAlert(h),
	}
}

// Severity returns the alert severity
func (tpa *TransferPausedAlert) Severity() Severity {
	return Info
}

// String returns string representation
func (tpa *TransferPausedAlert) String() string {
	return fmt.Sprintf("TransferPausedAlert{hash=%s}", tpa.Hash.String())
}

// TransferResumedAlert represents a transfer resumed alert
type TransferResumedAlert struct {
	*TransferAlert
}

// NewTransferResumedAlert creates a new transfer resumed alert
func NewTransferResumedAlert(h *hash.Hash) *TransferResumedAlert {
	return &TransferResumedAlert{
		TransferAlert: NewTransferAlert(h),
	}
}

// Severity returns the alert severity
func (tra *TransferResumedAlert) Severity() Severity {
	return Info
}

// String returns string representation
func (tra *TransferResumedAlert) String() string {
	return fmt.Sprintf("TransferResumedAlert{hash=%s}", tra.Hash.String())
}

// TransferRemovedAlert represents a transfer removed alert
type TransferRemovedAlert struct {
	*TransferAlert
}

// NewTransferRemovedAlert creates a new transfer removed alert
func NewTransferRemovedAlert(h *hash.Hash) *TransferRemovedAlert {
	return &TransferRemovedAlert{
		TransferAlert: NewTransferAlert(h),
	}
}

// Severity returns the alert severity
func (tra *TransferRemovedAlert) Severity() Severity {
	return Info
}

// String returns string representation
func (tra *TransferRemovedAlert) String() string {
	return fmt.Sprintf("TransferRemovedAlert{hash=%s}", tra.Hash.String())
}