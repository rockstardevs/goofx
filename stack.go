package goofx

import (
	"encoding/xml"
	"errors"
)

// TagStack is a stack of xml StartElements.
type TagStack interface {
	Push(*xml.StartElement)
	Pop() (*xml.StartElement, error)
	IsEmpty() bool
	Size() int
	Dump() []string
}

// stack is a stack of xml tags pointers.
type stack struct {
	items []*xml.StartElement
}

// NewStack returns an initialized empty stack.
func NewStack() TagStack {
	return &stack{
		items: make([]*xml.StartElement, 0),
	}
}

// Push adds the given element to top of stack.
func (s *stack) Push(t *xml.StartElement) {
	s.items = append(s.items, t)
}

// Pop removes and returns the topmost elment of the stack.
func (s *stack) Pop() (*xml.StartElement, error) {
	l := len(s.items)
	if l == 0 {
		return nil, errors.New("error - popping from empty stack")
	}
	i := s.items[l-1]
	s.items = s.items[:l-1]
	return i, nil
}

// IsEmpty returns true if the stack is empty, else false.
func (s *stack) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the current size of the stack.
func (s *stack) Size() int {
	return len(s.items)
}

// Dump returns a string representation of the stack for debugging.
func (s *stack) Dump() []string {
	result := make([]string, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item.Name.Local)
	}
	return result
}
