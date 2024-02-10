package wal

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/oklog/ulid"
)

// Segment is a collection of records that exist in the WAL, in the order they were
// appended. The WAL will append a maximum of maxSegmentSize records to a segment
// before cutting a new one.
type Segment[T any] struct {
	id      ulid.ULID
	records []T
}

func loadSegment[T any](id ulid.ULID, r io.Reader) (Segment[T], error) {
	result := Segment[T]{
		id: id,
	}

	if err := json.NewDecoder(r).Decode(&result.records); err != nil {
		return result, fmt.Errorf("failed to parse WAL segment: %w", err)
	}

	return result, nil
}

func (s *Segment[T]) serialize(w io.Writer) error {
	log.WithField("segment", s.id.String()).Tracef("Serializing Segment of size %d", s.Length())
	return json.NewEncoder(w).Encode(s.records)
}

func (s *Segment[T]) append(v T) {
	log.WithField("segment", s.id.String()).Tracef("Appending record: %+v", v)
	s.records = append(s.records, v)
}

func (s *Segment[T]) Length() int {
	return len(s.records)
}

// Records returns a copy of records in this segment
func (s *Segment[T]) Records() []T {
	result := make([]T, len(s.records))
	copy(result, s.records)

	return result
}
