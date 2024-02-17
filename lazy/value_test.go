package lazy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	var reset bool
	sut := New[string](func() {
		reset = true
	})

	assert.False(t, reset, "cache was reset unexpectedly")

	assert.Equal(t, "a", sut.Fetch(func() string {
		return "a"
	}))

	assert.False(t, reset, "cache was reset unexpectedly")

	assert.Equal(t, "a", sut.Fetch(func() string {
		return "b"
	}))

	assert.False(t, reset, "cache was reset unexpectedly")

	sut.Zero()
	assert.True(t, reset, "expected the cache to be reset")

	assert.Zero(t, sut.Fetch(func() string {
		return "c"
	}))
}
