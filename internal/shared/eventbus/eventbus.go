package eventbus

import (
	"context"
	"sync"
)

// Handler is a function that handles events
type Handler func(context.Context, interface{})

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu        sync.RWMutex
	handlers  map[string][]Handler
	eventChan chan eventWrapper
	quit      chan struct{}
	wg        sync.WaitGroup
}

type eventWrapper struct {
	ctx   context.Context
	event interface{}
	topic string
}

// New creates a new EventBus
func New(bufferSize int) *EventBus {
	bus := &EventBus{
		handlers:  make(map[string][]Handler),
		eventChan: make(chan eventWrapper, bufferSize),
		quit:      make(chan struct{}),
	}

	bus.wg.Add(1)
	go bus.processEvents()

	return bus
}

// Subscribe registers a handler for a specific topic
func (b *EventBus) Subscribe(topic string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = append(b.handlers[topic], handler)
}

// Publish sends an event to all subscribers of the topic
func (b *EventBus) Publish(ctx context.Context, topic string, event interface{}) {
	select {
	case b.eventChan <- eventWrapper{ctx: ctx, event: event, topic: topic}:
	case <-ctx.Done():
	case <-b.quit:
	}
}

// PublishAsync publishes an event without blocking
func (b *EventBus) PublishAsync(ctx context.Context, topic string, event interface{}) {
	go b.Publish(ctx, topic, event)
}

// processEvents handles event distribution to subscribers
func (b *EventBus) processEvents() {
	defer b.wg.Done()

	for {
		select {
		case wrapper := <-b.eventChan:
			b.dispatchEvent(wrapper)
		case <-b.quit:
			return
		}
	}
}

// dispatchEvent sends an event to all registered handlers for a topic
func (b *EventBus) dispatchEvent(wrapper eventWrapper) {
	b.mu.RLock()
	handlers := b.handlers[wrapper.topic]
	b.mu.RUnlock()

	for _, handler := range handlers {
		handler(wrapper.ctx, wrapper.event)
	}
}

// Close shuts down the event bus
func (b *EventBus) Close() {
	close(b.quit)
	b.wg.Wait()
	close(b.eventChan)
}

// Topics for the application
const (
	TopicTestEvent     = "test.event"
	TopicTestStarted   = "test.started"
	TopicTestCompleted = "test.completed"
	TopicTestFailed    = "test.failed"
	TopicPackageFound  = "package.found"
	TopicFSChanged     = "fs.changed"
	TopicCoverageReady = "coverage.ready"
	TopicError         = "error"
)
