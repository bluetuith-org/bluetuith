package eventbus

import (
	"sync"

	"github.com/cskr/pubsub/v2"
)

// NilEventHandler represents a disabled event handler.
type NilEventHandler struct{}

// DefaultEventHandler represents an internal event handler.
type DefaultEventHandler struct {
	*pubsub.PubSub[uint, any]
}

// EventPublisher represents an interface that provides an event publisher.
type EventPublisher interface {
	// Publish publishes an event to the event stream.
	Publish(id uint, data any)
}

// EventSubscriber represents an interface that provides an event subscriber.
type EventSubscriber interface {
	// Subscribe subscribes to an event from the event stream.
	Subscribe(id uint) SubscriberID
}

// EventHandler represents an interface that provides an event publisher and subscriber.
type EventHandler interface {
	EventPublisher
	EventSubscriber
}

// eventHandler represents the main event handler.
type eventHandler struct {
	p EventPublisher
	s EventSubscriber

	mu sync.RWMutex
}

var eventEmitter eventHandler

func init() {
	RegisterEventHandler(DefaultHandler())
}

// RegisterEventHandler registers the event handler interface.
func RegisterEventHandler[H EventHandler](eh H) {
	eventEmitter.mu.Lock()
	defer eventEmitter.mu.Unlock()

	eventEmitter.p = eh
	eventEmitter.s = eh
}

// RegisterEventHandlers registers the event publisher and subscriber interfaces separately.
// To disable an EventPublisher or EventSubscriber, pass 'nil' as the parameter.
// For example: `RegisterEventHandlers(&eventPublisher{}, nil)` can be called to only register
// an event publisher.
func RegisterEventHandlers[P EventPublisher, S EventSubscriber](p P, s S) {
	eventEmitter.mu.Lock()
	defer eventEmitter.mu.Unlock()

	eventEmitter.p = p
	eventEmitter.s = s
}

// DisableEvents unregisters the event handler.
func DisableEvents() {
	RegisterEventHandler(&NilEventHandler{})
}

// Publish calls the registered publisher handler.
func Publish(id EventID, data any) {
	if id == nil {
		return
	}

	eventEmitter.mu.RLock()
	p := eventEmitter.p
	eventEmitter.mu.RUnlock()

	p.Publish(id.Value(), data)
}

// Subscribe calls the registered subscriber handler.
func Subscribe(id EventID) SubscriberID {
	if id == nil {
		return (&NilEventHandler{}).Subscribe(0)
	}

	eventEmitter.mu.RLock()
	s := eventEmitter.s
	eventEmitter.mu.RUnlock()

	return s.Subscribe(id.Value())
}

// DefaultHandler returns the default event handler.
func DefaultHandler() *DefaultEventHandler {
	return &DefaultEventHandler{PubSub: pubsub.New[uint, any](10)}
}

// NilHandler returns a disabled event handler.
func NilHandler() *NilEventHandler {
	return &NilEventHandler{}
}

// Publish publishes an event to the event stream.
func (d *DefaultEventHandler) Publish(id uint, data any) {
	d.TryPub(data, id)
}

// Subscribe subscribes to an event from the event stream.
func (d *DefaultEventHandler) Subscribe(id uint) SubscriberID {
	ch := d.Sub(id)
	return SubscriberID{
		C:      ch,
		active: true,
		unsub: func() {
			go d.Unsub(ch, id)
		},
	}
}

// Publish does not do anything.
func (n *NilEventHandler) Publish(uint, any) {
}

// Subscribe does not do anything.
func (n *NilEventHandler) Subscribe(uint) SubscriberID {
	ch := make(chan any)
	close(ch)
	return SubscriberID{C: ch}
}
