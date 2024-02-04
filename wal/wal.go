package wal

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid"
)

var ulidEntropySource = ulid.Monotonic(
	// We don't have to be cryptographically secure, and each scrobble should yield at most
	// one new segment, so using math/rand is fine here.
	rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	0,
)

// WAL is a simple write-ahead-log like structure containing multiple records. It is
// rooted at a specific directory, which contains WAL segments as files. Each file is
// named with a ULID for the segment based on the timestamp the segment was created.
// Segments contain a maximum of maxSegmentSize records, and segments are automatically
// added as new records are appended to the log, as well as removed as segments are
// successfully processed.
type WAL[T any] struct {
	root     string
	segments []*Segment[T]

	maxSegmentSize int
}

// Open constructs a new WAL by loading segments from the specified directory. Directories
// and files that are not named with a valid ULID are ignored. Each segment is loaded
// in order based on the timestamp the segment was created, which is embedded in the ID
// of the segment.
func Open[T any](path string, maxSegmentSize int) (WAL[T], error) {
	w := WAL[T]{root: path, maxSegmentSize: maxSegmentSize}

	// Ensure the WAL directory exists
	if err := os.MkdirAll(w.root, 0700); err != nil {
		return w, fmt.Errorf("open WAL: failed to ensure WAL directory: %w", err)
	}

	// Get a listing of the WAL directory contents
	files, err := os.ReadDir(path)
	if err != nil {
		return w, fmt.Errorf("open WAL: failed to list WAL segments: %w", err)
	}

	// ULIDs are already in order and os.ReadDir returns the listing in order
	for _, file := range files {
		// Skip directories, we only care about files
		if file.IsDir() {
			continue
		}

		id, err := ulid.ParseStrict(file.Name())
		if err != nil {
			// Not a WAL segment
			continue
		}

		// Open and load the segment
		f, err := os.OpenFile(filepath.Join(w.root, file.Name()), os.O_RDONLY|os.O_SYNC, 0600)
		if err != nil {
			return w, fmt.Errorf("open WAL: failed to open segment %s: %w", id.String(), err)
		}

		segment, err := loadSegment[T](id, f)
		if err != nil {
			return w, fmt.Errorf("open WAL: failed to load segment %s: %w", id.String(), err)
		}

		w.segments = append(w.segments, &segment)
	}

	return w, nil
}

// Append adds the specified value to the WAL, creating a new segment if required. Once
// the record is appended to the tail of the WAL, the segment containing it is committed
// to disk.
func (w *WAL[T]) Append(v T) error {
	// Create a segment if we don't have any or the current tail segment is out of space
	if len(w.segments) == 0 || w.segments[len(w.segments)-1].length() >= w.maxSegmentSize {
		w.segments = append(w.segments, &Segment[T]{
			id: ulid.MustNew(ulid.Now(), ulidEntropySource),
		})
	}

	tail := w.segments[len(w.segments)-1]

	tail.append(v)

	f, err := os.OpenFile(filepath.Join(w.root, tail.id.String()), os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_SYNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to commit append: failed to open WAL Segment: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	return tail.serialize(f)
}

// Process iterates through WAL segments in order and invokes the specified visitation
// function on each segment. If the function returns no error, the segment is trimmed
// from the WAL. If the function returns an error, processing stops and the segment is
// retained.
func (w *WAL[T]) Process(visit func(segment Segment[T]) error) error {
	for len(w.segments) > 0 {
		head := w.segments[0]

		if err := visit(*head); err != nil {
			return fmt.Errorf("process WAL segment %s: %w", head.id.String(), err)
		}

		// Remove the segment from the filesystem
		if err := os.Remove(filepath.Join(w.root, head.id.String())); err != nil {
			return fmt.Errorf("process WAL segment %s: failed to trim segment: %w", head.id.String(), err)
		}

		// Remove the segment from the list
		w.segments = w.segments[1:]
	}

	return nil
}
