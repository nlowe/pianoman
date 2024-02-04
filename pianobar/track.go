package pianobar

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	keyArtist       = "artist"
	keyTitle        = "title"
	keyAlbum        = "album"
	keySongDuration = "songDuration"
	keySongPlayed   = "songPlayed"
	keyRating       = "rating"
)

type Track struct {
	Artist string
	Title  string
	Album  string

	ThumbsUp bool

	SongDuration time.Duration
	SongPlayed   time.Duration

	ScrobbleAt time.Time
}

func TrackFromReader(r io.Reader) (Track, error) {
	lines := bufio.NewScanner(r)

	var result Track
	result.ScrobbleAt = time.Now().UTC()

	for lines.Scan() {
		if err := lines.Err(); err != nil {
			return result, err
		}

		parts := strings.SplitN(lines.Text(), "=", 2)
		intPart, _ := strconv.Atoi(parts[1])

		switch parts[0] {
		case keyArtist:
			result.Artist = parts[1]
		case keyTitle:
			result.Title = parts[1]
		case keyAlbum:
			result.Album = parts[1]
		case keyRating:
			result.ThumbsUp = parts[1] == "1"
		case keySongDuration:
			result.SongDuration = time.Duration(intPart) * time.Second
		case keySongPlayed:
			result.SongPlayed = time.Duration(intPart) * time.Second
		}
	}

	return result, nil
}
