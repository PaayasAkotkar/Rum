// Package rumstack
// Simple Stack implementation
package rumstack

import (
	"slices"
	"sync"
)

// Stack -> new entry at first, old entry at last
type Stack[T comparable] struct {
	mu   sync.Mutex
	data []T
}

func NewStack[T comparable]() *Stack[T] {
	return &Stack[T]{}
}

// get funcs
func (s *Stack[T]) len() int    { return len(s.data) }
func (s *Stack[T]) empty() bool { return len(s.data) == 0 }

func (s *Stack[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.len()
}

func (s *Stack[T]) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.empty()
}

// Latest returns a copy of the newest entry
func (s *Stack[T]) Latest() *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.empty() {
		return nil
	}
	v := s.data[0]
	return &v
}

// Oldest returns a copy of the oldest entry
func (s *Stack[T]) Oldest() *T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.empty() {
		return nil
	}
	v := s.data[len(s.data)-1]
	return &v
}

func (s *Stack[T]) Max() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]T, s.len())
	copy(out, s.data)
	return out
}

// Range returns a copy of up to limit entries
func (s *Stack[T]) Range(limit int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.empty() || limit > s.len() {
		return nil
	}
	out := make([]T, limit)
	copy(out, s.data[:limit])
	return out
}

// end

// set funcs

// Push inserts at the front (newest first)
func (s *Stack[T]) Push(name T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append([]T{name}, s.data...)
}

func (s *Stack[T]) Erase(name T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.data {
		if r == name {
			s.data = append(s.data[:i], s.data[i+1:]...)
			return
		}
	}
}

// Rearrange moves the [from, to] range to the front
func (s *Stack[T]) Rearrange(from T, to T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fIdx := slices.Index(s.data, from)
	tIdx := slices.Index(s.data, to)
	if fIdx == -1 || tIdx == -1 || fIdx > tIdx {
		return
	}

	priority := make([]T, tIdx-fIdx+1)
	copy(priority, s.data[fIdx:tIdx+1])

	remaining := make([]T, 0, len(s.data)-(tIdx-fIdx+1))
	remaining = append(remaining, s.data[:fIdx]...)
	remaining = append(remaining, s.data[tIdx+1:]...)

	s.data = append(priority, remaining...)
}

// end
