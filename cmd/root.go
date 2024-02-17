package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/nlowe/pianoman/lazy"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nlowe/pianoman/internal/config"
	"github.com/nlowe/pianoman/lastfm"
	"github.com/nlowe/pianoman/pianobar"
	"github.com/nlowe/pianoman/pianobar/eventcmd"
	"github.com/nlowe/pianoman/wal"
)

func NewRootCmd() *cobra.Command {
	var cfg config.Config

	result := &cobra.Command{
		Use:   "pianoman <eventcmd>",
		Short: "A simple Last.FM Scrobbler for pianobar",
		Long:  "pianoman scrobbles events to Last.FM by integrating with pianobar's eventcmd interface",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			u, err := user.Current()
			if err != nil {
				return fmt.Errorf("failed to identify current user: %w", err)
			}

			configFilePath := filepath.Join(u.HomeDir, ".config/pianoman/config.yaml")
			f, err := os.Open(configFilePath)
			if err != nil {
				return fmt.Errorf("failed to open config: %w", err)
			}

			defer func() {
				_ = f.Close()
			}()

			mode, err := f.Stat()
			if err == nil && mode.Mode().Perm() != 0o600 {
				logrus.Warnf("Config file has insecure permissions. Want: 600, got %o", mode.Mode().Perm())
			}

			cfg, err = config.Parse(f)
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}

			cfg.Path = configFilePath

			// Update Logger
			lvl, err := logrus.ParseLevel(cfg.Verbosity)
			if err != nil {
				return fmt.Errorf("failed to configure logging verbosity: %w", err)
			}
			logrus.SetLevel(lvl)

			return err
		},
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			// Normalize WAL Path
			w, err := wal.Open[pianobar.Track](filepath.Join(
				filepath.Dir(cfg.Path), filepath.Clean(cfg.Scrobble.WALDirectory),
			), lastfm.MaxTracksPerScrobble)

			if err != nil {
				return fmt.Errorf("failed to open wal: %w", err)
			}

			sessionTokenCachePath := filepath.Join(
				filepath.Dir(cfg.Path), "session",
			)

			sessionTokenCache := lazy.New[string](func() {
				logrus.Debug("Deleting Session Token")
				_ = os.Remove(sessionTokenCachePath)
			})

			defer func() {
				token := sessionTokenCache.Fetch(func() string {
					return ""
				})

				if token == "" {
					logrus.Warn("No token to cache")
					return
				}

				// Try to cache the token
				logrus.Debug("Caching session token")
				if err = os.WriteFile(sessionTokenCachePath, []byte(token), 0o600); err != nil {
					logrus.WithError(err).Error("Failed to cache session token")
				}
			}()

			cachedToken, err := os.ReadFile(sessionTokenCachePath)
			if err == nil {
				logrus.Debug("Using cached session token")
				_ = sessionTokenCache.Fetch(func() string {
					return strings.TrimSpace(string(cachedToken))
				})
			}

			lfm := lastfm.New(
				sessionTokenCache,
				cfg.Auth.API.Key,
				cfg.Auth.API.Secret,
				cfg.Auth.User.Name,
				cfg.Auth.User.Password,
			)

			flags := eventcmd.HandleSongFinish
			if cfg.Scrobble.NowPlaying {
				flags |= eventcmd.HandleSongStart
			}

			if cfg.Scrobble.Thumbs {
				flags |= eventcmd.HandleSongLove
				flags |= eventcmd.HandleSongBan
			}

			// TODO: Don't scrobble thumbs down if configured

			// TODO: Support eventcmd chaining
			_, err = eventcmd.Handle(ctx, args[0], flags, os.Stdin, w, lfm, lfm)
			return err
		},
	}

	return result
}
