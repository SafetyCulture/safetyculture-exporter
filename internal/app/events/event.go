package events

type eventType string

const (
	EventTypeError        eventType = "TYPE_ERROR"
	EventTypeFeedProgress eventType = "TYPE_FEED_PROGRESS"
)

type Event struct {
	EventType eventType   `json:"event_type"`
	Error     *EventError `json:"error,omitempty"`
	FeedInfo  *FeedInfo   `json:"feed_info,omitempty"`
}

func (e *Event) IsErrorType() bool {
	return e.EventType == EventTypeError
}

func (e *Event) IsFeedInfoType() bool {
	return e.EventType == EventTypeFeedProgress
}
