package events

import (
	"fmt"

	"go.uber.org/zap"
)

type errorSeverity string
type errorSubSystem string

const (
	ErrorSeverityInfo    errorSeverity = "INFO"
	ErrorSeverityWarning               = "WARNING"
	ErrorSeverityError                 = "ERROR"
)

const (
	ErrorSubSystemDB             errorSubSystem = "DB"
	ErrorSubSystemDataIntegrity                 = "Data Integrity"
	ErrorSubSystemAPI                           = "API"
	ErrorSubSystemFileOperations                = "File Operations"
)

type EventError struct {
	error
	severity  errorSeverity
	subsystem errorSubSystem
	isFatal   bool
}

func (ee *EventError) IsInfo() bool {
	return ee.severity == ErrorSeverityInfo
}

func (ee *EventError) IsWarn() bool {
	return ee.severity == ErrorSeverityWarning
}

func (ee *EventError) IsError() bool {
	return ee.severity == ErrorSeverityError
}

func (ee *EventError) IsFatal() bool {
	return ee.isFatal
}

func (ee *EventError) Log(log *zap.SugaredLogger) {
	switch ee.severity {
	case ErrorSeverityError:
		log.Errorf("%s:%s", ee.error, ee.subsystem)
	case ErrorSeverityWarning:
		log.Warnf("%s:%s", ee.error, ee.subsystem)
	default:
		log.Infof("%s:%s", ee.error, ee.subsystem)
	}
}

func BuildNewEventError(severity errorSeverity, subsystem errorSubSystem, fatal bool, err error) *Event {
	return &Event{
		EventType: EventTypeError,
		Error: &EventError{
			error:     err,
			severity:  severity,
			subsystem: subsystem,
			isFatal:   fatal,
		},
		FeedInfo: nil,
	}
}

func NewErrorEventInfo(err error, message string) *EventError {
	return &EventError{
		error:    err,
		severity: ErrorSeverityInfo,
		isFatal:  false,
	}
}

func NewErrorEventWarning(err error, message string) *EventError {
	return &EventError{
		error:    err,
		severity: ErrorSeverityWarning,
		isFatal:  false,
	}
}

func NewErrorEventError(err error, message string) error {
	return &EventError{
		error:    err,
		severity: ErrorSeverityError,
		isFatal:  false,
	}
}

// NewEventError creates a new EventError
func NewEventError(err error, severity errorSeverity, subsystem errorSubSystem, fatal bool) error {
	return &EventError{
		error:     err,
		severity:  severity,
		subsystem: subsystem,
		isFatal:   fatal,
	}
}

// NewEventErrorWithMessage creates a new EventError
func NewEventErrorWithMessage(err error, severity errorSeverity, subsystem errorSubSystem, fatal bool, msg string) error {
	evErr := &EventError{
		error:     err,
		severity:  severity,
		subsystem: subsystem,
		isFatal:   fatal,
	}
	return WrapEventError(evErr, msg)
}

// WrapEventError wraps error
func WrapEventError(err error, message string) error {
	newErr := fmt.Errorf("%s: %w", message, err)
	switch theError := err.(type) {
	case *EventError:
		theError.error = newErr
		return theError

	default:
		return newErr
	}
}
