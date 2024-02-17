package lastfm

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"

	"github.com/nlowe/pianoman/pianobar"
)

var log = logrus.WithField("prefix", "lastfm")

const (
	apiRoot = "https://ws.audioscrobbler.com/2.0"

	// https://www.last.fm/api/mobileauth
	// https://www.last.fm/api/show/auth.getMobileSession
	methodGetMobileSession = "auth.getMobileSession"
	// https://www.last.fm/api/show/track.scrobble
	methodScrobble = "track.scrobble"
	// https://www.last.fm/api/show/track.updateNowPlaying
	methodUpdateNowPlaying = "track.updateNowPlaying"
	// https://www.last.fm/api/show/track.love
	methodLoveTrack = "track.love"
	// https://www.last.fm/api/show/track.unlove
	methodUnLoveTrack = "track.unlove"

	// MaxTracksPerScrobble is the maximum number of tracks that can be included in a single request to Scrobble
	MaxTracksPerScrobble = 50
)

// Scrobbler sends track information to the Last.FM API as Scrobbles. It also provides a way to notify Last.FM of the
// track a user is currently listening to. If Scrobble returns a non-terminal error, the provided track is saved and
// re-tried on the next scrobble. The Last.FM API accepts up to 50 tracks in one scrobble request.
type Scrobbler interface {
	// Scrobble emits the specified tracks to Last.FM's track.Scrobble API. Tracks must meet the following criteria to be
	// accepted by Last.FM:
	//
	// * The track must be at least 30s long
	// * The track must have been played for at least half of its duration, or for at least 4 minutes
	//
	// These criteria should be checked by the caller before calling Scrobble.
	//
	// Any non-terminal error should save the tracks to be retried on the next scrobble request. Up to 50 tracks may be
	// queued for a single request. If the provided context is cancelled before the request can be made to Last.FM's API
	// it will be retried on the next request. If it is cancelled after the request has been sent but before the
	// response can be read, it will not be retried later.
	Scrobble(ctx context.Context, t ...pianobar.Track) error

	// UpdateNowPlaying submits the specified track to Last.FM's track.updateNowPlaying API
	UpdateNowPlaying(ctx context.Context, t pianobar.Track) error
}

// FeedbackProvider provides a way to translate pandora feedback on tracks to Last.FM
type FeedbackProvider interface {
	// LoveTrack should be called for tracks that have received a Thumbs-Up from the user
	LoveTrack(ctx context.Context, t pianobar.Track) error
	// UnLoveTrack should be called for tracks that have been banned or un-loved by the user
	UnLoveTrack(ctx context.Context, t pianobar.Track) error
}

type API struct {
	api *http.Client

	getSessionKey *sync.Once

	sessionKey string
	apiKey     string
	apiSecret  string
	username   string
	password   string
}

// Ensure API implements Scrobbler and FeedbackProvider
var _ Scrobbler = (*API)(nil)
var _ FeedbackProvider = (*API)(nil)

func New(key, secret, username, password string) *API {
	return &API{
		api: cleanhttp.DefaultClient(),

		getSessionKey: &sync.Once{},

		apiKey:    key,
		apiSecret: secret,
		username:  username,
		password:  password,
	}
}

func sendAndCheck[TResult any](ctx context.Context, a *API, params Request) (Response[TResult], error) {
	var result Response[TResult]

	// Sign the request
	log.Debugf("Signing %s request", params.method())
	params.sign(a.apiKey, a.apiSecret, a.sessionKey)

	// Send the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiRoot, nil)
	if err != nil {
		return result, fmt.Errorf("sendAndCheck: failed to build request: %w", err)
	}

	req.URL.RawQuery = params.encode()

	resp, err := a.api.Do(req)
	if err != nil {
		return result, fmt.Errorf("sendAndCheck: failed to make request: %w", err)
	}

	// Decode Response
	log.Tracef("%s finished with %s", params.method(), resp.Status)
	if err = xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("sendAndCheck: request failed: failed to parse response: %s: %w", resp.Status, err)
	}

	// Check Response
	log.Tracef("Last.FM returned %s in response to %s", result.Status, params.method())
	if result.Status == statusFailed {
		return result, fmt.Errorf("sendAndCheck: request failed: %s: %w", resp.Status, result.Error)
	}

	if result.Status != statusOK {
		return result, fmt.Errorf("sendAndCheck: request failed: unknown status (%s) %s", resp.Status, result.Status)
	}

	return result, nil
}

func (a *API) ensureSessionKey(ctx context.Context) {
	a.getSessionKey.Do(func() {
		// TODO: Cache session key?

		log.Debugf("Logging into Last.FM as %s", a.username)
		params := newRequest(methodGetMobileSession)
		params.set(paramApiKey, a.apiKey)
		params.set("username", a.username)
		params.set("password", a.password)

		resp, err := sendAndCheck[Session](ctx, a, params)
		if err != nil {
			log.WithError(err).Fatal("Failed to login to Last.FM")
		}

		a.sessionKey = resp.Value.Key
	})
}

// Scrobble sends the provided track and all other pending scrobbles to https://www.last.fm/api/show/track.scrobble
func (a *API) Scrobble(ctx context.Context, tracks ...pianobar.Track) error {
	if len(tracks) == 0 {
		return fmt.Errorf("scrobble: must provide at least one track")
	}

	if len(tracks) > MaxTracksPerScrobble {
		return fmt.Errorf("scrobble: up to %d tracks may be included in one scrobble request: got %d", MaxTracksPerScrobble, len(tracks))
	}

	a.ensureSessionKey(ctx)

	log.Debugf("Scrobbling %d track(s)", len(tracks))

	// Populate Tracks
	params := newRequest(methodScrobble)
	for i, t := range tracks {
		params.set(fmt.Sprintf("artist[%d]", i), t.Artist)
		params.set(fmt.Sprintf("track[%d]", i), t.Title)
		params.set(fmt.Sprintf("timestamp[%d]", i), strconv.Itoa(int(t.ScrobbleAt.Unix())))
		params.set(fmt.Sprintf("album[%d]", i), t.Album)
		params.set(fmt.Sprintf("chosenByUser[%d]", i), "0")
		params.set(fmt.Sprintf("duration[%d]", i), strconv.Itoa(int(t.SongDuration.Seconds())))
	}

	resp, err := sendAndCheck[ScrobbleResult](ctx, a, params)
	log.Tracef("Last.FM accepted %d track(s) and ignored %d track(s)", resp.Value.Accepted, resp.Value.Ignored)

	return err
}

// UpdateNowPlaying calls https://www.last.fm/api/show/track.updateNowPlaying
func (a *API) UpdateNowPlaying(ctx context.Context, t pianobar.Track) error {
	a.ensureSessionKey(ctx)

	log.Debugf("Updating now-playing: %+v", t)

	// Populate Track Data
	params := newRequest(methodUpdateNowPlaying)

	params.set("artist", t.Artist)
	params.set("track", t.Title)
	params.set("album", t.Album)
	params.set("duration", strconv.Itoa(int(t.SongDuration.Seconds())))

	_, err := sendAndCheck[Track](ctx, a, params)
	return err
}

// LoveTrack calls https://www.last.fm/api/show/track.love
func (a *API) LoveTrack(ctx context.Context, t pianobar.Track) error {
	a.ensureSessionKey(ctx)

	log.Debugf("Loving Track: %+v", t)
	params := newRequest(methodLoveTrack)

	params.set("artist", t.Artist)
	params.set("track", t.Title)

	_, err := sendAndCheck[struct{}](ctx, a, params)
	return err
}

// UnLoveTrack calls https://www.last.fm/api/show/track.unlove
func (a *API) UnLoveTrack(ctx context.Context, t pianobar.Track) error {
	a.ensureSessionKey(ctx)

	log.Debugf("Un-Loving Track: %+v", t)

	params := newRequest(methodUnLoveTrack)

	params.set("artist", t.Artist)
	params.set("track", t.Title)

	_, err := sendAndCheck[struct{}](ctx, a, params)

	return err
}
