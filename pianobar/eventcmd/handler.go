package eventcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nlowe/pianoman/lastfm"
	"github.com/nlowe/pianoman/pianobar"
	"github.com/nlowe/pianoman/wal"
)

var log = logrus.WithField("prefix", "handler")

type EventFlags uint8

const (
	EventSongStart  = "songstart"
	EventSongFinish = "songfinish"
	EventSongLove   = "songlove"
	EventSongBan    = "songban"

	HandleSongStart EventFlags = 1 << iota
	HandleSongFinish
	HandleSongLove
	HandleSongBan
)

func (e EventFlags) checkEventAndFlags(event, desired string, flag EventFlags) bool {
	return strings.EqualFold(desired, event) && (e&flag == flag)
}

// ShouldHandle returns true iff the correct flag is set for the specified event
func (e EventFlags) ShouldHandle(event string) bool {
	return e.checkEventAndFlags(event, EventSongStart, HandleSongStart) ||
		e.checkEventAndFlags(event, EventSongFinish, HandleSongFinish) ||
		e.checkEventAndFlags(event, EventSongLove, HandleSongLove) ||
		e.checkEventAndFlags(event, EventSongBan, HandleSongBan)
}

// Handle processes a command executed by pianobar's eventcmd interface. First, it checks to see if the provided event
// is handled based on the set EventFlags. If not, it returns the passed reader and no error.
//
// If the event is enabled, the stdin reader is read fully and the event is handled. Then, a new reader is returned as
// well as any error from the handling of the event.
//
// In either case, if event chaining is enabled, the returned reader can be used to re-read the eventcmd payload.
func Handle(
	ctx context.Context,
	event string,
	handle EventFlags,
	stdin io.Reader,
	w wal.WAL[pianobar.Track],
	s lastfm.Scrobbler,
	f lastfm.FeedbackProvider,
) (io.Reader, error) {
	log.Debugf("Received event: %s", event)
	if !handle.ShouldHandle(event) {
		log.Trace("Ignoring event due to flags")
		return stdin, nil
	}

	// Save the event payload, so it can be read again by chained commands
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read eventcmd payload: %w", err)
	}

	payload := string(raw)
	log.Tracef("Received event payload: \n%s\n", payload)

	next := strings.NewReader(payload)

	// All of the events we could handle use a track as the payload
	track, err := pianobar.TrackFromReader(strings.NewReader(payload))
	if err != nil {
		return next, fmt.Errorf("failed to parse track from eventcmd payload: %w", err)
	}

	log = log.WithFields(logrus.Fields{
		"artist": track.Artist,
		"album":  track.Album,
		"title":  track.Title,
	})

	// Dispatch the event
	switch event {
	case EventSongStart:
		log.Info("Updating Now Playing")
		err = s.UpdateNowPlaying(ctx, track)
	case EventSongFinish:
		log.Info("Scrobbling Track")
		err = handleFinish(ctx, track, w, s)
	case EventSongLove:
		log.Info("Sending feedback to Last.FM")
		err = f.LoveTrack(ctx, track)
	case EventSongBan:
		log.Info("Sending feedback to Last.FM")
		// Last.FM doesn't have a ban/block, the best we can do is un-love
		err = f.UnLoveTrack(ctx, track)
	default:
		err = fmt.Errorf("unknown event: %s", event)
	}

	// And return the saved reader and any error from event handling
	return next, err
}

func handleFinish(ctx context.Context, t pianobar.Track, w wal.WAL[pianobar.Track], s lastfm.Scrobbler) error {
	// Check if we've met the requirements for a scrobble
	//
	// From: https://www.last.fm/api/scrobbling#when-is-a-scrobble-a-scrobble
	// * The track must be longer than 30 seconds.
	// * And the track has been played for at least half its duration, or for 4 minutes (whichever occurs earlier.)
	if t.SongDuration < 30*time.Second {
		return nil
	}

	if !(t.SongPlayed > 4*time.Minute || (float64(t.SongPlayed)/float64(t.SongDuration)) > 0.5) {
		return nil
	}

	// Append the track to the WAL in case of an error
	if err := w.Append(t); err != nil {
		return fmt.Errorf("failed to append track to WAL: %w", err)
	}

	// Try to scrobble the WAL Backlog
	return w.Process(func(segment wal.Segment[pianobar.Track]) error {
		err := s.Scrobble(ctx, segment.Records()...)

		// Only retry invalid session key, service offline, and service unavailable
		// Additionally, retry Operation Failed and Rate Limit Exceeded
		// Any other errors (i.e. from the network stack) will be retried.
		lfm := &lastfm.Error{}
		if err != nil && errors.As(err, &lfm) {
			if !slices.Contains([]int{8, 9, 11, 16, 29}, lfm.Code) {
				// This error shouldn't be retried according to the LFM Docs
				err = nil
			}
		}

		if err != nil {
			err = fmt.Errorf("failed to scrobble tracks: %w", err)
		}

		return err
	})
}
