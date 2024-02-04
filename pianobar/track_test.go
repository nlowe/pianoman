package pianobar

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackFromReader(t *testing.T) {
	sut, err := TrackFromReader(strings.NewReader(`artist=Test Artist
title=Test Title with=foo
album=Test Album
coverArt=https://nlowe.dev/foo.jpg
stationName=Some Station
songStationName=Some Other Station
pRet=0
pRetStr=OK
wRet=0
wRetStr=OK
songDuration=456
songPlayed=123
rating=1
detailUrl=https://nlowe.dev/bar.json
`))

	require.NoError(t, err)

	assert.Equal(t, "Test Artist", sut.Artist)
	assert.Equal(t, "Test Title with=foo", sut.Title)
	assert.Equal(t, "Test Album", sut.Album)
	assert.True(t, sut.ThumbsUp)

	assert.EqualValues(t, 456*time.Second, sut.SongDuration)
	assert.EqualValues(t, 123*time.Second, sut.SongPlayed)
}
