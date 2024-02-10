# pianoman

[![](https://github.com/nlowe/pianoman/workflows/CI/badge.svg)](https://github.com/nlowe/pianoman/actions) [![Coverage Status](https://coveralls.io/repos/github/nlowe/pianoman/badge.svg?branch=master)](https://coveralls.io/github/nlowe/pianoman?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/nlowe/pianoman)](https://goreportcard.com/report/github.com/nlowe/pianoman) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](./LICENSE)

A [Last.FM](https://www.last.fm/) Scrobbler for [pianobar](https://github.com/PromyLOPh/pianobar)

```
|>  Station "Rock / Metal" (445386914731145560)
(i) Receiving new playlist... Ok.
...
|>  "Best Part" by "The Score" on "Carry On" <3
[2024-02-10T15:51:09-05:00] TRACE wal: Opening wall at '/home/nlowe/.config/pianoman/wal' with max segment size 50
[2024-02-10T15:51:09-05:00] DEBUG lastfm: Logging into Last.FM as nlowe0
[2024-02-10T15:51:09-05:00] DEBUG lastfm: Signing auth.getMobileSession request
[2024-02-10T15:51:10-05:00] TRACE lastfm: auth.getMobileSession finished with 200 OK
[2024-02-10T15:51:10-05:00] TRACE lastfm: Last.FM returned ok in response to auth.getMobileSession
[2024-02-10T15:51:10-05:00] DEBUG handler: Received event: songstart
[2024-02-10T15:51:10-05:00] TRACE handler: Received event payload:
artist=The Score
title=Best Part
album=Carry On
coverArt=http://mediaserver-cont-dc6-1-v4v6.pandora.com/images/03/d6/7f/e1/e5904c89ae0e6e040d850c99/1080W_1080H.jpg
stationName=Rock / Metal
songStationName=
pRet=1
pRetStr=Everything is fine :)
wRet=0
wRetStr=No error
songDuration=178
songPlayed=0
rating=1
detailUrl=http://www.pandora.com/score/carry-on/best-part/TRg72nw3p349Vmm?dc=1777&ad=0:27:1:44039::0:0:0:0:510:019:OH:39093:0:0:0:0:0:0
stationCount=2
station0=QuickMix
station1=Rock / Metal


[2024-02-10T15:51:10-05:00] DEBUG lastfm: Updating now-playing: {Artist:The Score Title:Best Part Album:Carry On ThumbsUp:true SongDuration:2m58s SongPlayed:0s ScrobbleAt:2024-02-10 20:51:10.258500291 +0000 UTC}
[2024-02-10T15:51:10-05:00] DEBUG lastfm: Signing track.updateNowPlaying request
[2024-02-10T15:51:11-05:00] TRACE lastfm: track.updateNowPlaying finished with 200 OK
[2024-02-10T15:51:11-05:00] TRACE lastfm: Last.FM returned ok in response to track.updateNowPlaying
...
[2024-02-10T15:54:10-05:00] TRACE wal: Opening wall at '/home/nlowe/.config/pianoman/wal' with max segment size 50
[2024-02-10T15:54:10-05:00] DEBUG lastfm: Logging into Last.FM as nlowe0
[2024-02-10T15:54:10-05:00] DEBUG lastfm: Signing auth.getMobileSession request
[2024-02-10T15:54:10-05:00] TRACE lastfm: auth.getMobileSession finished with 200 OK
[2024-02-10T15:54:10-05:00] TRACE lastfm: Last.FM returned ok in response to auth.getMobileSession
[2024-02-10T15:54:10-05:00] DEBUG handler: Received event: songfinish
[2024-02-10T15:54:10-05:00] TRACE handler: Received event payload:
artist=The Score
title=Best Part
album=Carry On
coverArt=http://mediaserver-cont-dc6-1-v4v6.pandora.com/images/03/d6/7f/e1/e5904c89ae0e6e040d850c99/1080W_1080H.jpg
stationName=Rock / Metal
songStationName=
pRet=1
pRetStr=Everything is fine :)
wRet=0
wRetStr=No error
songDuration=177
songPlayed=177
rating=1
detailUrl=http://www.pandora.com/score/carry-on/best-part/TRg72nw3p349Vmm?dc=1777&ad=0:27:1:44039::0:0:0:0:510:019:OH:39093:0:0:0:0:0:0
stationCount=2
station0=QuickMix
station1=Rock / Metal


[2024-02-10T15:54:10-05:00] TRACE wal: Creating new segment 01HPACS430DZAP6RAGETJM1Z4K
[2024-02-10T15:54:10-05:00]  INFO wal: Appending record: {Artist:The Score Title:Best Part Album:Carry On ThumbsUp:true SongDuration:2m57s SongPlayed:2m57s ScrobbleAt:2024-02-10 20:54:10.783966194 +0000 UTC} segment=01HPACS430DZAP6RAGETJM1Z4K
[2024-02-10T15:54:10-05:00] TRACE wal: Serializing Segment of size 1 segment=01HPACS430DZAP6RAGETJM1Z4K
[2024-02-10T15:54:10-05:00] DEBUG wal: Processing WAL
[2024-02-10T15:54:10-05:00] TRACE wal: Processing segment segment=01HPACS430DZAP6RAGETJM1Z4K
[2024-02-10T15:54:10-05:00] DEBUG lastfm: Scrobbling 1 track(s)
[2024-02-10T15:54:10-05:00] DEBUG lastfm: Signing track.scrobble request
[2024-02-10T15:54:10-05:00] TRACE lastfm: track.scrobble finished with 200 OK
[2024-02-10T15:54:10-05:00] TRACE lastfm: Last.FM returned ok in response to track.scrobble
[2024-02-10T15:54:10-05:00] TRACE lastfm: Last.FM accepted %d track(s) and ignored %d track(s)1 0
[2024-02-10T15:54:10-05:00] TRACE wal: Successfully Processed segment, attempting  to trim segment=01HPACS430DZAP6RAGETJM1Z4K
...
```

## Requirements

Follow [the Last.FM Documentation](https://www.last.fm/api/authentication) to get an API Key and secret.

Update `~/.config/pianoman/config.yaml` with the API Key and your `Last.FM` credentials.

Then, set `event_command` to point at `pianoman` and start listening!

## Configuration

Place the following config template in `~/.config/pianoan/config.yaml`. Because this config contains secrets,
you should ensure only your user is able to read it. Treat it as you would an SSH key.

```yaml
auth:
  # Last.FM API Key and secret
  # See https://www.last.fm/api/authentication
  api:
    key: '***'
    secret: '***'
  # Your Last.FM Credentials
  user:
    name: someone@gmail.com
    password: '***'

scrobble:
  # Update the user's currently playing track
  nowPlaying: true
  # Mark any thumbs-up'd tracks as loved and any banned tracks as un-loved
  thumbs: true
  # Don't scrobble any tracks that are thumbs-down'd
  ignoreThumbsDown: true
  # Where to store the scrobble log. Each segment contains up
  # to 50 tracks to scrobble. Each segment is sent as one batch
  # and is retried as one batch. Scrobbles that are filtered are
  # not retried.
  #
  # This directory is relative to the config file
  wal: 'wal'

# Chain the eventcmd metadata to another program (including
# events that aren't handled by pianoman). If specified, this
# program will be invoked exactly like pianoman was, regardless
# of whether or not pianoman was able to successfully scrobble
# tracks.
#eventcmd:
#  next: '/opt/pianobar/notify.py'

# The level to log at. One of:
# trace, debug, info, warning, error, fatal, off.
verbosity: info
```
