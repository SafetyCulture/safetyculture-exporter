package events

const (
	ErrorSeverityInfo    errorSeverity = "INFO"
	ErrorSeverityWarning               = "WARNING"
	ErrorSeverityError                 = "ERROR"
)

type EventError struct {
	error
	severity        errorSeverity
	isFatal         bool
	SimpleMessage   string `json:"simple_message"`
	DetailedMessage string `json:"detailed_message"`
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

func BuildEventError(severity errorSeverity, fatal bool, simpleMessage string, detailedMessage string, err error) *Event {
	return &Event{
		EventType: EventTypeError,
		Error: &EventError{
			error:           err,
			severity:        severity,
			isFatal:         fatal,
			SimpleMessage:   simpleMessage,
			DetailedMessage: detailedMessage,
		},
		FeedInfo: nil,
	}
}
