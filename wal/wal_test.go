package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nlowe/pianoman/lastfm"
)

func TestWAL(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(root, "dir"), 0700))
	_, err := os.Create(filepath.Join(root, "not_a_segment"))
	require.NoError(t, err)

	sut, err := Open[int](root, lastfm.MaxTracksPerScrobble)
	require.NoError(t, err)

	// Generate some test values, should create two segments
	for i := 0; i < 75; i++ {
		require.NoError(t, sut.Append(i))
	}

	require.Len(t, sut.segments, 2)

	// Re-Open the WAL to exercise segment loading
	sut, err = Open[int](root, lastfm.MaxTracksPerScrobble)
	require.NoError(t, err)
	require.Len(t, sut.segments, 2)

	// Process the first segment and stop
	var notFirst bool
	require.ErrorContains(t, sut.Process(func(segment Segment[int]) error {
		if notFirst {
			return fmt.Errorf("dummy")
		}

		notFirst = true

		records := segment.Records()
		require.Len(t, records, lastfm.MaxTracksPerScrobble)

		for i := 0; i < lastfm.MaxTracksPerScrobble; i++ {
			assert.Equalf(t, i, records[i], "element at position %d does not have expected value: got %d", i, records[i])
		}

		return nil
	}), "dummy")

	// We should have one segment left
	require.Len(t, sut.segments, 1)

	// Re-Open the WAL again to verify a single segment
	sut, err = Open[int](root, lastfm.MaxTracksPerScrobble)
	require.NoError(t, err)
	require.Len(t, sut.segments, 1)

	require.NoError(t, sut.Process(func(segment Segment[int]) error {
		records := segment.Records()
		require.Len(t, records, 25)

		for i := 0; i < 25; i++ {
			assert.Equalf(t, lastfm.MaxTracksPerScrobble+i, records[i], "element at position %d does not have expected value: got %d", i, records[i])
		}

		return nil
	}))

	// Re-Open the WAL one last time to verify no segments remain
	sut, err = Open[int](root, lastfm.MaxTracksPerScrobble)
	require.NoError(t, err)
	require.Empty(t, sut.segments)
}
