# pianoman

[![](https://github.com/nlowe/pianoman/workflows/CI/badge.svg)](https://github.com/nlowe/pianoman/actions) [![Coverage Status](https://coveralls.io/repos/github/nlowe/pianoman/badge.svg?branch=master)](https://coveralls.io/github/nlowe/pianoman?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/nlowe/pianoman)](https://goreportcard.com/report/github.com/nlowe/pianoman) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](./LICENSE)

A [Last.FM](https://www.last.fm/) Scrobbler for [pianobar](https://github.com/PromyLOPh/pianobar)

Very Early WIP. Still TODO:

* Logging
* Config
* Wire up the Root Command

## Requirements

Follow [the Last.FM Documentation](https://www.last.fm/api/authentication) to get an API Key and secret.

Update `~/.config/pianoman/config.yaml` with the API Key and your `Last.FM` credentials.

## Configuration

Place the following config template in `~/.config/pianoan/config.yaml`. Because this config contains secrets,
you should ensure only your user is able to read it. Treat it as you would an SSH key.

```yaml
auth:
  # Last.FM API Key and secret
  # See https://www.last.fm/api/authentication
  api:
    key: '***'
    secret: '**'
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
