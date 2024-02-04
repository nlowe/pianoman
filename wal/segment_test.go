package wal

import (
	"strings"
	"testing"

	"github.com/oklog/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegment(t *testing.T) {
	t.Run("loadSegment", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			src := strings.NewReader(`[1,2,3]`)

			id := ulid.MustNew(ulid.Now(), ulidEntropySource)
			sut, err := loadSegment[int](id, src)
			require.NoError(t, err)

			assert.Equal(t, id, sut.id, "segment id")

			records := sut.Records()
			assert.Len(t, records, 3, "record count")
			assert.Equal(t, 1, records[0])
			assert.Equal(t, 2, records[1])
			assert.Equal(t, 3, records[2])
		})

		t.Run("error", func(t *testing.T) {
			_, err := loadSegment[int](ulid.MustNew(ulid.Now(), ulidEntropySource), strings.NewReader("definitely not json"))

			require.ErrorContains(t, err, "failed to parse WAL segment:")
		})
	})

	t.Run("serialize", func(t *testing.T) {
		sut := Segment[int]{records: []int{1, 2, 3}}

		buff := &strings.Builder{}
		require.NoError(t, sut.serialize(buff))

		assert.Equal(t, "[1,2,3]\n", buff.String())
	})

	t.Run("append", func(t *testing.T) {
		sut := Segment[int]{records: []int{1, 2, 3}}

		assert.Equal(t, 3, sut.length())

		sut.append(4)

		assert.Equal(t, 4, sut.length())

		records := sut.Records()
		assert.Equal(t, 4, records[3])
	})
}
