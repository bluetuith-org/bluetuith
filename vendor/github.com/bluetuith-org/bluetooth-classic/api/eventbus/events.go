package eventbus

// EventID represents a unique event ID.
type EventID interface {
	String() string
	Value() uint
}

// UnsubFunc describes a function to be called when unsubscribing from an event.
type UnsubFunc func()

// SubscriberID represents a subcriber ID.
type SubscriberID struct {
	C      chan any
	active bool
	unsub  UnsubFunc
}

// Unsubscribe unsubscribes from the attached subscription.
func (s SubscriberID) Unsubscribe() {
	if s.unsub != nil {
		s.unsub()
	}
}

// IsActive returns if the subscriber can actually receive events.
func (s SubscriberID) IsActive() bool {
	return s.active
}
