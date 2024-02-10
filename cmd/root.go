package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"

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
	var w wal.WAL[pianobar.Track]
	var lfm *lastfm.API

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

			// TODO: Warn on insecure config mode bits

			defer func() {
				_ = f.Close()
			}()

			cfg, err = config.Parse(f)
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}

			// Update Logger
			lvl, err := logrus.ParseLevel(cfg.Verbosity)
			if err != nil {
				return fmt.Errorf("failed to configure logging verbosity: %w", err)
			}
			logrus.SetLevel(lvl)

			// Normalize WAL Path
			w, err = wal.Open[pianobar.Track](filepath.Join(
				filepath.Dir(configFilePath), filepath.Clean(cfg.Scrobble.WALDirectory),
			), lastfm.MaxTracksPerScrobble)

			return err
		},
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			var err error
			lfm, err = lastfm.New(
				ctx,
				cfg.Auth.API.Key,
				cfg.Auth.API.Secret,
				cfg.Auth.User.Name,
				cfg.Auth.User.Password,
			)
			if err != nil {
				return fmt.Errorf("failed to authenticate to Last.FM: %w", err)
			}

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
