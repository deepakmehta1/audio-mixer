package service

import "sync"

// Broadcaster handles broadcasting data to multiple subscribers.
type Broadcaster struct {
	subscribers map[chan []byte]struct{}
	lock        sync.Mutex
}

// NewBroadcaster creates a new broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan []byte]struct{}),
	}
}

// Subscribe returns a new channel that will receive broadcasted data.
func (b *Broadcaster) Subscribe() chan []byte {
	ch := make(chan []byte, 64)
	b.lock.Lock()
	b.subscribers[ch] = struct{}{}
	b.lock.Unlock()
	return ch
}

// Unsubscribe removes a channel from the broadcaster.
func (b *Broadcaster) Unsubscribe(ch chan []byte) {
	b.lock.Lock()
	delete(b.subscribers, ch)
	b.lock.Unlock()
}

// Broadcast sends data to all subscribers.
func (b *Broadcaster) Broadcast(data []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	for ch := range b.subscribers {
		select {
		case ch <- data:
		default:
			// Skip slow subscribers.
		}
	}
}

// Global broadcaster instance.
var GlobalBroadcaster = NewBroadcaster()
