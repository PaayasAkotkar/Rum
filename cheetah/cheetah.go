package cheetah

import (
	"sync"
)

// Cheetah  is a generic pub/sub system keyed by string.
// T is the message payload type.
type Cheetah[T any] struct {
	mu          sync.Mutex
	subscribers map[string]map[chan *T]struct{}
}

// New returns new Cheetah  of single buffer
// single for fast process 😄
func New[T any]() *Cheetah[T] {
	return &Cheetah[T]{
		subscribers: make(map[string]map[chan *T]struct{}, 1),
	}
}

// Subscribe returns a buffered channel that receives published values for key
func (l *Cheetah[T]) Subscribe(key string) chan *T {
	ch := make(chan *T, 1)
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.subscribers[key]; !ok {
		l.subscribers[key] = make(map[chan *T]struct{}, 1)
	}
	l.subscribers[key][ch] = struct{}{}
	return ch
}

// Publish sends results to all subscribers of key. Never blocks.
func (l *Cheetah[T]) Publish(key string, parcel *T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ch := range l.subscribers[key] {
		select {
		case ch <- parcel:
		default:
		}
	}
}

// Unsubscribe removes and closes the channel for key
func (l *Cheetah[T]) Unsubscribe(key string, body chan *T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if sub, ok := l.subscribers[key]; ok {
		delete(sub, body)
		close(body)
		if len(sub) == 0 {
			delete(l.subscribers, key)
		}
	}
}
