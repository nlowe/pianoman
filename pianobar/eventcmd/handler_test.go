package eventcmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nlowe/pianoman/internal/fake"
	"github.com/nlowe/pianoman/lastfm"
	"github.com/nlowe/pianoman/pianobar"
	"github.com/nlowe/pianoman/wal"
)

const defaultTestTrack = `artist=Test Artist
title=Test Title
album=Test Album
songDuration=300
songPlayed=175`

func isDefaultTestTrack(v any) bool {
	vt := v.(pianobar.Track)

	return vt.Album == "Test Album" && vt.Artist == "Test Artist" && vt.Title == "Test Title"
}

func setup(t *testing.T) (w wal.WAL[pianobar.Track], s *fake.Scrobbler, f *fake.FeedbackProvider) {
	t.Helper()
	d := t.TempDir()

	w, err := wal.Open[pianobar.Track](d, lastfm.MaxTracksPerScrobble)
	require.NoError(t, err)

	return w, fake.NewScrobbler(t), fake.NewFeedbackProvider(t)
}

func invokeExpecting(t *testing.T, errHandler func(require.TestingT, error, ...any), event string, flags EventFlags, payload string, w wal.WAL[pianobar.Track], s *fake.Scrobbler, f *fake.FeedbackProvider) {
	t.Helper()
	next, err := Handle(context.Background(), event, flags, strings.NewReader(payload), w, s, f)
	errHandler(t, err)

	v, err := io.ReadAll(next)
	require.NoError(t, err)
	require.Equal(t, payload, string(v))
}

func invoke(t *testing.T, event string, flags EventFlags, payload string, w wal.WAL[pianobar.Track], s *fake.Scrobbler, f *fake.FeedbackProvider) {
	t.Helper()

	invokeExpecting(t, require.NoError, event, flags, payload, w, s, f)
}

func TestHandler_songstart(t *testing.T) {
	w, s, f := setup(t)

	s.EXPECT().UpdateNowPlaying(mock.Anything, mock.MatchedBy(isDefaultTestTrack)).Return(nil)

	invoke(t, EventSongStart, HandleSongStart, defaultTestTrack, w, s, f)
}

func TestHandler_songfinish(t *testing.T) {
	t.Run("Too Short", func(t *testing.T) {
		w, s, f := setup(t)
		payload := `artist=Test Artist
title=Test Title
album=Test Album
songDuration=15
songPlayed=15`

		invoke(t, EventSongFinish, HandleSongFinish, payload, w, s, f)
	})

	t.Run("Not Enough Played", func(t *testing.T) {
		w, s, f := setup(t)
		payload := `artist=Test Artist
title=Test Title
album=Test Album
songDuration=300
songPlayed=15`

		invoke(t, EventSongFinish, HandleSongFinish, payload, w, s, f)
	})

	t.Run("Accept", func(t *testing.T) {
		t.Run("Long Enough", func(t *testing.T) {
			w, s, f := setup(t)
			payload := `artist=Test Artist
title=Test Title
album=Test Album
songDuration=600
songPlayed=270`

			s.EXPECT().Scrobble(mock.Anything, mock.MatchedBy(isDefaultTestTrack)).Return(nil)

			invoke(t, EventSongFinish, HandleSongFinish, payload, w, s, f)
		})

		t.Run("Half Played", func(t *testing.T) {
			w, s, f := setup(t)

			s.EXPECT().Scrobble(mock.Anything, mock.MatchedBy(func(v any) bool {
				vt := v.(pianobar.Track)

				return vt.Album == "Test Album" && vt.Artist == "Test Artist" && vt.Title == "Test Title"
			})).Return(nil)

			invoke(t, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
		})

		t.Run("Error", func(t *testing.T) {
			// Test errors that shouldn't be retried. The handler should not return an error
			t.Run("Terminal", func(t *testing.T) {
				w, s, f := setup(t)

				s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(&lastfm.Error{
					Code:    7,
					Message: "Invalid resource specified",
				})

				invoke(t, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
			})

			t.Run("Retry", func(t *testing.T) {
				t.Run("Generic Errors", func(t *testing.T) {
					w, s, f := setup(t)

					s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(fmt.Errorf("dummy"))

					invokeExpecting(t, func(t require.TestingT, err error, _ ...any) {
						require.ErrorContains(t, err, "dummy")
					}, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
				})

				t.Run("LastFM Errors", func(t *testing.T) {
					t.Run("OperationFailed", func(t *testing.T) {
						w, s, f := setup(t)

						s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(&lastfm.Error{
							Code:    8,
							Message: "Operation failed - Most likely the backend service failed. Please try again.",
						})

						invokeExpecting(t, func(t require.TestingT, err error, _ ...any) {
							require.ErrorContains(t, err, "8: Operation failed")
						}, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
					})

					t.Run("InvalidSessionKey", func(t *testing.T) {
						w, s, f := setup(t)

						s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(&lastfm.Error{
							Code:    9,
							Message: "Invalid session key - Please re-authenticate",
						})

						invokeExpecting(t, func(t require.TestingT, err error, _ ...any) {
							require.ErrorContains(t, err, "9: Invalid session key")
						}, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
					})

					t.Run("ServiceOffline", func(t *testing.T) {
						w, s, f := setup(t)

						s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(&lastfm.Error{
							Code:    11,
							Message: "Service Offline - This service is temporarily offline. Try again later.",
						})

						invokeExpecting(t, func(t require.TestingT, err error, _ ...any) {
							require.ErrorContains(t, err, "11: Service Offline")
						}, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
					})

					t.Run("ServiceUnavailable", func(t *testing.T) {
						w, s, f := setup(t)

						s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(&lastfm.Error{
							Code:    16,
							Message: "The service is temporarily unavailable, please try again.",
						})

						invokeExpecting(t, func(t require.TestingT, err error, _ ...any) {
							require.ErrorContains(t, err, "16: The service is temporarily unavailable")
						}, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
					})

					t.Run("RateLimitExceeded", func(t *testing.T) {
						w, s, f := setup(t)

						s.EXPECT().Scrobble(mock.Anything, mock.Anything).Return(&lastfm.Error{
							Code:    29,
							Message: "Rate Limit Exceded - Your IP has made too many requests in a short period, exceeding our API guidelines",
						})

						invokeExpecting(t, func(t require.TestingT, err error, _ ...any) {
							require.ErrorContains(t, err, "29: Rate Limit Exceded")
						}, EventSongFinish, HandleSongFinish, defaultTestTrack, w, s, f)
					})
				})
			})
		})
	})
}

func TestHandler_songlove(t *testing.T) {
	w, s, f := setup(t)

	f.EXPECT().LoveTrack(mock.Anything, mock.MatchedBy(isDefaultTestTrack)).Return(nil)

	invoke(t, EventSongLove, HandleSongLove, defaultTestTrack, w, s, f)
}

func TestHandler_songban(t *testing.T) {
	w, s, f := setup(t)

	f.EXPECT().UnLoveTrack(mock.Anything, mock.MatchedBy(isDefaultTestTrack)).Return(nil)

	invoke(t, EventSongBan, HandleSongBan, defaultTestTrack, w, s, f)
}
