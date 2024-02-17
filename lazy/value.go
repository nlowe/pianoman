package lazy

import (
	"sync"
)

// Value is a simple wrapper around sync.Once that can lazy load a value
// Once Zero is called, this value always returns the zero value for T.
type Value[T any] struct {
	once *sync.Once

	value T

	onZero func()
}

// New creates a new Value. The provided onZero function will be called when Zero is called
func New[T any](onZero func()) *Value[T] {
	return &Value[T]{
		once:   &sync.Once{},
		onZero: onZero,
	}
}

// Fetch returns the lazy value, calling populate to fetch it the first time
func (c *Value[T]) Fetch(populate func() T) T {
	c.once.Do(func() {
		c.value = populate()
	})

	return c.value
}

// Zero causes any future calls to Fetch to return the zero value for this function
func (c *Value[T]) Zero() {
	var v T
	c.value = v
	c.onZero()
}
