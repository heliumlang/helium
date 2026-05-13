package util

import (
	"errors"
	"slices"
)

type Set[T comparable] struct {
	slice []T
}

func (s *Set[T]) Push(v T) error {
	if slices.Contains(s.slice, v) {
		return errors.New("element already in set")
	}

	s.slice = append(s.slice, v)
	return nil
}

func (s *Set[T]) Pop() (*T, error) {
	if len(s.slice) == 0 {
		return nil, errors.New("empty slice")
	}

	v := s.slice[len(s.slice)-1]
	s.slice = s.slice[:len(s.slice)-1]
	return &v, nil
}

func (s *Set[T]) Remove(i int) error {
	if i > len(s.slice)-1 || i < 0 {
		return errors.New("index out of bounds")
	}

	s.slice = append(s.slice[:i], s.slice[i+1:]...)
	return nil
}

func (s *Set[T]) IndexOf(target T) (int, error) {
	for i, v := range s.slice {
		if v == target {
			return i, nil
		}
	}

	return -1, errors.New("not found")
}
