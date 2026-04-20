package chakra

import (
	"log"
	"sync"
)

// Chakra is a generic pub/sub system keyed by string.
// T is the message payload type.
type Chakra[T any] struct {
	mu          sync.Mutex
	subscribers map[string]map[chan *T]struct{}
}

// New returns new chakra of single buffer
// single for fast process 😄
func New[T any]() *Chakra[T] {
	return &Chakra[T]{
		subscribers: make(map[string]map[chan *T]struct{}, 1),
	}
}

// Subscribe returns a buffered channel that receives published values for key
func (l *Chakra[T]) Subscribe(key string) chan *T {
	ch := make(chan *T, 1)
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.subscribers[key]; !ok {
		l.subscribers[key] = make(map[chan *T]struct{}, 1)
	}
	l.subscribers[key][ch] = struct{}{}
	return ch
}

// KageBunshinNoJutsu sends results to all subscribers of key. Never blocks.
// KageBunshinNoJutsu or MultiShadowCloneJutsu or publish
func (l *Chakra[T]) KageBunshinNoJutsu(key string, val *T) {
	log.Println("in publish")
	l.mu.Lock()
	defer l.mu.Unlock()
	for ch := range l.subscribers[key] {
		select {
		case ch <- val:
		default:
		}
	}
}

// Kai removes and closes the channel for key
// Kai or dispell or unsub
func (l *Chakra[T]) Kai(key string, ch chan *T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if sub, ok := l.subscribers[key]; ok {
		delete(sub, ch)
		close(ch)
		if len(sub) == 0 {
			delete(l.subscribers, key)
		}
	}
}
