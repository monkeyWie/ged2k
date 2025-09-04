package exception

import "fmt"

// ErrorCode represents different error types in the ed2k system
type ErrorCode int

const (
	NoError ErrorCode = iota
	ServerConnUnsupportedPacket
	PeerConnUnsupportedPacket
	PacketHeaderUndefined
	InflateError
	PacketSizeIncorrect
	PacketSizeOverflow
	ServerMetHeaderIncorrect
	GenericInstantiationError
	GenericIllegalAccess
	EndOfStream
	IOError
	NoTransfer
	FileNotFound
	OutOfParts
	ConnectionTimeout
	ChannelClosed
	QueueRanking
	FileIOError
	UnableToDeleteFile
	InternalError
	BufferUnderflowException
	BufferGetException
	WrongHashset
	HashMismatch
	NonWriteableChannel
)

const (
	TagTypeUnknown = 30 + iota
	TagToStringInvalid
	TagToIntInvalid
	TagToLongInvalid
	TagToFloatInvalid
	TagToHashInvalid
	TagFromStringInvalidCP
	TagToBlobInvalid
	TagToBsobInvalid
)

const (
	DuplicatePeer = 40 + iota
	DuplicatePeerConnection
	PeerLimitExceeded
	SecurityException
	UnsupportedEncoding
	IllegalArgument
)

const (
	TransferFinished = 50 + iota
	TransferPaused
	TransferAborted
)

const (
	NoMemory = 60 + iota
	SessionStopping
	IncomingDirInaccessible
	BufferTooLarge
	NotConnected
	Interrupted
)

const (
	PortMappingAlreadyMapped = 70 + iota
	PortMappingNoDevice
	PortMappingError
	PortMappingIOError
	PortMappingSAXError
	PortMappingConfigError
	PortMappingException
	PortMappingCommandRejected
)

const (
	DHTRequestAlreadyRunning = 80 + iota
	DHTTrackerAborted
)

const (
	LinkMalformed = 90 + iota
	URISyntaxError
	NumberFormatError
	UnknownLinkType
	GithubCfgIPIsNull
	GithubCfgPortsAreNull
	GithubCfgPortsAreEmpty
	InvalidPRParameter
	PeerRequestOverflow
)

const (
	Fail = 100
)

// Error messages for each error code
var errorMessages = map[ErrorCode]string{
	NoError:                      "No error",
	ServerConnUnsupportedPacket:  "Server unsupported packet",
	PeerConnUnsupportedPacket:    "Peer connection unsupported packet",
	PacketHeaderUndefined:        "Packet header contains wrong bytes or undefined",
	InflateError:                 "Inflate error",
	PacketSizeIncorrect:          "Packet size less than zero",
	PacketSizeOverflow:           "Packet size too big",
	ServerMetHeaderIncorrect:     "Server met file contains incorrect header byte",
	GenericInstantiationError:    "Generic instantiation error",
	GenericIllegalAccess:         "Generic illegal access",
	EndOfStream:                  "End of stream",
	IOError:                      "I/O exception",
	NoTransfer:                   "No transfer",
	FileNotFound:                 "File not found",
	OutOfParts:                   "Out of parts",
	ConnectionTimeout:            "Connection timeout",
	ChannelClosed:                "Channel closed",
	QueueRanking:                 "Queue ranking",
	FileIOError:                  "File I/O error occurred",
	UnableToDeleteFile:           "Unable to delete file",
	InternalError:                "Internal product error",
	BufferUnderflowException:     "Buffer underflow exception",
	BufferGetException:           "Buffer get method raised common exception",
	WrongHashset:                 "Wrong hash set",
	HashMismatch:                 "Hash mismatch",
	NonWriteableChannel:          "Non writeable channel",
	TagTypeUnknown:               "Tag type unknown",
	TagToStringInvalid:           "Tag to string conversion error",
	TagToIntInvalid:              "Tag to int conversion error",
	TagToLongInvalid:             "Tag to long conversion error",
	TagToFloatInvalid:            "Tag to float conversion error",
	TagToHashInvalid:             "Tag to hash conversion error",
	TagFromStringInvalidCP:       "Tag from string creation error invalid code page",
	TagToBlobInvalid:             "Tag to blob conversion error",
	TagToBsobInvalid:             "Tag to bsob conversion error",
	DuplicatePeer:                "Duplicate peer",
	DuplicatePeerConnection:      "Duplicate peer connection",
	PeerLimitExceeded:            "Peer limit exceeded",
	SecurityException:            "Security exception",
	UnsupportedEncoding:          "Unsupported encoding exception",
	IllegalArgument:              "Illegal argument",
	TransferFinished:             "Transfer finished",
	TransferPaused:               "Transfer paused",
	TransferAborted:              "Transfer aborted",
	NoMemory:                     "No memory available",
	SessionStopping:              "Session stopping",
	IncomingDirInaccessible:      "Incoming directory is inaccessible",
	BufferTooLarge:               "Buffer too large",
	NotConnected:                 "Not connected",
	Interrupted:                  "Interrupted",
	PortMappingAlreadyMapped:     "Port already mapped",
	PortMappingNoDevice:          "No gateway device found",
	PortMappingError:             "Unable to map port",
	PortMappingIOError:           "I/O exception on mapping port",
	PortMappingSAXError:          "SAX parsing exception on port mapping",
	PortMappingConfigError:       "Configuration exception on port mapping",
	PortMappingException:         "Unknown exception on port mapping",
	PortMappingCommandRejected:   "Mapping command was rejected",
	DHTRequestAlreadyRunning:     "DHT request with the same hash already in progress",
	DHTTrackerAborted:            "DHT tracker was already aborted at the moment",
	LinkMalformed:                "Incorrect link format",
	URISyntaxError:               "URI has incorrect syntax",
	NumberFormatError:            "Parse number exception",
	UnknownLinkType:              "Emule link has unrecognized type",
	GithubCfgIPIsNull:            "IP is null in github kad config",
	GithubCfgPortsAreNull:        "Ports are null in github kad config",
	GithubCfgPortsAreEmpty:       "Ports are empty in github kad config",
	InvalidPRParameter:           "Peer request parameters are invalid",
	PeerRequestOverflow:          "Peer request has length greater than PIECE_SIZE",
	Fail:                         "Fail",
}

// String returns the error description
func (e ErrorCode) String() string {
	if msg, ok := errorMessages[e]; ok {
		return fmt.Sprintf("%s {%d}", msg, int(e))
	}
	return fmt.Sprintf("Unknown error {%d}", int(e))
}

// Code returns the numeric error code
func (e ErrorCode) Code() int {
	return int(e)
}

// Description returns just the error description without code
func (e ErrorCode) Description() string {
	if msg, ok := errorMessages[e]; ok {
		return msg
	}
	return "Unknown error"
}

// JED2KException represents an exception in the ed2k system
type JED2KException struct {
	ErrorCode ErrorCode
	Message   string
}

// NewJED2KException creates a new exception with error code
func NewJED2KException(code ErrorCode) *JED2KException {
	return &JED2KException{
		ErrorCode: code,
		Message:   code.Description(),
	}
}

// NewJED2KExceptionWithMessage creates a new exception with error code and custom message
func NewJED2KExceptionWithMessage(code ErrorCode, message string) *JED2KException {
	return &JED2KException{
		ErrorCode: code,
		Message:   message,
	}
}

// Error implements the error interface
func (e *JED2KException) Error() string {
	return fmt.Sprintf("JED2K Error [%d]: %s", e.ErrorCode.Code(), e.Message)
}