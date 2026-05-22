package util

import "fmt"

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Pop() T {
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v
}

func (s *Stack[T]) Push(v T) int {
	s.items = append(s.items, v)
	return len(s.items) - 1
}

func (s *Stack[T]) Get(i int) (*T, error) {
	if i < 0 || i > len(s.items) {
		return nil, fmt.Errorf("index out of bounds")
	}

	v := s.items[i]
	return &v, nil
}
