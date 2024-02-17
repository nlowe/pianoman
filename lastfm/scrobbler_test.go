package lastfm

import (
	"context"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nlowe/pianoman/lazy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nlowe/pianoman/pianobar"
)

const (
	testSessionKey = "secret"
	testApiKey     = "123"
	testApiSecret  = "abc"
)

type roundTripperFunc func(r *http.Request) *http.Response

func (r roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	resp := r(request)

	if resp.Header == nil {
		// Header must not be nil
		resp.Header = http.Header{}
	}

	return resp, nil
}

func setupAPI(t *testing.T, f func(r *http.Request) *http.Response) *API {
	t.Helper()

	tokenCache := lazy.New[string](func() {})

	_ = tokenCache.Fetch(func() string {
		return testSessionKey
	})

	return &API{
		api: &http.Client{Transport: roundTripperFunc(f)},

		sessionKeyCache: tokenCache,

		apiKey:    testApiKey,
		apiSecret: testApiSecret,
	}
}

func assertHasParam(t *testing.T, params url.Values, k, v string) {
	t.Helper()

	assert.Equalf(t, v, params.Get(k), "key %s with value %s not specified in %s", k, v, params.Encode())
}

func assertAuthenticatedSignedRequest(t *testing.T, params url.Values, method string) {
	t.Helper()

	assertHasParam(t, params, "method", method)
	assertHasParam(t, params, "api_key", testApiKey)
	assertHasParam(t, params, "sk", testSessionKey)

	sigcheck := url.Values{}
	maps.Copy(sigcheck, params)
	delete(sigcheck, "api_sig")

	// TODO: This just checks that we signed the request the same way, not that we did it correctly
	Request(sigcheck).sign(testApiKey, testApiSecret, testSessionKey)
	assertHasParam(t, params, "api_sig", sigcheck.Get("api_sig"))
}

func TestAPI_Scrobble(t *testing.T) {
	t.Run("No Tracks", func(t *testing.T) {
		sut := setupAPI(t, func(r *http.Request) *http.Response {
			t.Error("No Request should have been made")
			return &http.Response{}
		})

		assert.EqualError(
			t,
			sut.Scrobble(context.Background()),
			"scrobble: must provide at least one track",
		)
	})

	t.Run("Too Many Tracks", func(t *testing.T) {
		sut := setupAPI(t, func(r *http.Request) *http.Response {
			t.Error("No Request should have been made")
			return &http.Response{}
		})

		payload := make([]pianobar.Track, 51)
		assert.EqualError(
			t,
			sut.Scrobble(context.Background(), payload...),
			"scrobble: up to 50 tracks may be included in one scrobble request: got 51",
		)
	})

	t.Run("OK", func(t *testing.T) {
		sut := setupAPI(t, func(r *http.Request) *http.Response {
			t.Helper()

			params := r.URL.Query()

			assertAuthenticatedSignedRequest(t, params, "track.scrobble")
			assertHasParam(t, params, "artist[0]", "Test Artist 0")
			assertHasParam(t, params, "track[0]", "Test Track 0")
			assertHasParam(t, params, "timestamp[0]", "1287141093")
			assertHasParam(t, params, "artist[1]", "Test Artist 1")
			assertHasParam(t, params, "track[1]", "Test Track 1")
			assertHasParam(t, params, "timestamp[1]", "1287141093")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`<?xml version='1.0' encoding='utf-8'?>
<lfm status="ok">
  <scrobbles accepted="2" ignored="0">
    <scrobble>
      <track corrected="0">Test Track 0</track>
      <artist corrected="0">Test Artist 0</artist>
      <album corrected="0"></album>
      <albumArtist corrected="0"></albumArtist>
      <timestamp>1287141093</timestamp>
      <ignoredMessage code="0"></ignoredMessage>
    </scrobble>
    <scrobble>
      <track corrected="0">Test Track 1</track>
      <artist corrected="0">Test Artist 1</artist>
      <album corrected="0"></album>
      <albumArtist corrected="0"></albumArtist>
      <timestamp>1287141093</timestamp>
      <ignoredMessage code="0"></ignoredMessage>
    </scrobble>
  </scrobbles>
</lfm>`)),
			}
		})

		require.NoError(t, sut.Scrobble(
			context.Background(),
			pianobar.Track{Title: "Test Track 0", Artist: "Test Artist 0", ScrobbleAt: time.Unix(1287141093, 0)},
			pianobar.Track{Title: "Test Track 1", Artist: "Test Artist 1", ScrobbleAt: time.Unix(1287141093, 0)},
		))
	})
}

func TestAPI_UpdateNowPlaying(t *testing.T) {
	sut := setupAPI(t, func(r *http.Request) *http.Response {
		t.Helper()

		params := r.URL.Query()

		assertAuthenticatedSignedRequest(t, params, "track.updateNowPlaying")
		assertHasParam(t, params, "artist", "Bad Wolves")
		assertHasParam(t, params, "track", "NDA")
		assertHasParam(t, params, "album", "Die About It")
		assertHasParam(t, params, "duration", "313")

		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`<?xml version='1.0' encoding='utf-8'?>
<lfm status="ok">
  <nowplaying>
    <track corrected="0">NDA</track>
    <artist corrected="0">Bad Wolves</artist>
    <album corrected="0">Die About It</album>
    <albumArtist corrected="0"></albumArtist>
    <ignoredMessage code="0"></ignoredMessage>
  </nowplaying>
</lfm>`)),
		}
	})

	require.NoError(t, sut.UpdateNowPlaying(context.Background(), pianobar.Track{
		Artist:       "Bad Wolves",
		Title:        "NDA",
		Album:        "Die About It",
		SongDuration: 5*time.Minute + 13*time.Second,
	}))
}

func TestAPI_FeedbackProvider(t *testing.T) {
	t.Run("Love", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			sut := setupAPI(t, func(r *http.Request) *http.Response {
				t.Helper()

				params := r.URL.Query()

				assertAuthenticatedSignedRequest(t, params, "track.love")
				assertHasParam(t, params, "artist", "Bad Wolves")
				assertHasParam(t, params, "track", "NDA")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`<lfm status="ok">
</lfm>`)),
				}
			})

			require.NoError(t, sut.LoveTrack(context.Background(), pianobar.Track{
				Title:  "NDA",
				Artist: "Bad Wolves",
			}))
		})

		t.Run("error", func(t *testing.T) {
			sut := setupAPI(t, func(r *http.Request) *http.Response {
				t.Helper()

				params := r.URL.Query()

				assertAuthenticatedSignedRequest(t, params, "track.love")
				assertHasParam(t, params, "artist", "Bad Wolves")
				assertHasParam(t, params, "track", "NDA")

				return &http.Response{
					Status:     http.StatusText(http.StatusForbidden),
					StatusCode: http.StatusForbidden,
					Body: io.NopCloser(strings.NewReader(`<lfm status="failed">
    <error code="10">Invalid API Key</error>
</lfm>`)),
				}
			})

			require.EqualError(t, sut.LoveTrack(context.Background(), pianobar.Track{
				Title:  "NDA",
				Artist: "Bad Wolves",
			}), "sendAndCheck: request failed: Forbidden: Last.FM API Error Code 10: Invalid API Key")
		})
	})

	t.Run("UnLove", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			sut := setupAPI(t, func(r *http.Request) *http.Response {
				t.Helper()

				params := r.URL.Query()

				assertAuthenticatedSignedRequest(t, params, "track.unlove")
				assertHasParam(t, params, "artist", "Taylor Swift")
				assertHasParam(t, params, "track", "Shake It Off")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`<lfm status="ok">
</lfm>`)),
				}
			})

			require.NoError(t, sut.UnLoveTrack(context.Background(), pianobar.Track{
				Title:  "Shake It Off",
				Artist: "Taylor Swift",
			}))
		})

		t.Run("error", func(t *testing.T) {
			sut := setupAPI(t, func(r *http.Request) *http.Response {
				t.Helper()

				params := r.URL.Query()

				assertAuthenticatedSignedRequest(t, params, "track.unlove")
				assertHasParam(t, params, "artist", "Taylor Swift")
				assertHasParam(t, params, "track", "Shake It Off")

				return &http.Response{
					Status:     http.StatusText(http.StatusForbidden),
					StatusCode: http.StatusForbidden,
					Body: io.NopCloser(strings.NewReader(`<lfm status="failed">
    <error code="10">Invalid API Key</error>
</lfm>`)),
				}
			})

			require.EqualError(t, sut.UnLoveTrack(context.Background(), pianobar.Track{
				Title:  "Shake It Off",
				Artist: "Taylor Swift",
			}), "sendAndCheck: request failed: Forbidden: Last.FM API Error Code 10: Invalid API Key")
		})

	})
}
