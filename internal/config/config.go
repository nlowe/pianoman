package config

import (
	"io"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config is the configuration used by pianoman
type Config struct {
	Auth     AuthConfig     `yaml:"auth"`
	Scrobble ScrobbleConfig `yaml:"scrobble"`

	EventCMD EventConfig `yaml:"eventcmd"`

	Verbosity string `yaml:"verbosity"`

	Path string `yaml:"-"`
}

type AuthConfig struct {
	API  APICredentials `yaml:"api"`
	User User           `yaml:"user"`
}

type APICredentials struct {
	Key    string `yaml:"key"`
	Secret string `yaml:"secret"`
}

type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type ScrobbleConfig struct {
	NowPlaying       bool `yaml:"nowPlaying"`
	Thumbs           bool `yaml:"thumbs"`
	IgnoreThumbsDown bool `yaml:"ignoreThumbsDown"`

	WALDirectory string `yaml:"wal"`
}

type EventConfig struct {
	Next string `yaml:"next"`
}

var defaultConfig = Config{
	Scrobble: ScrobbleConfig{
		NowPlaying:       true,
		Thumbs:           true,
		IgnoreThumbsDown: true,
		WALDirectory:     "wal",
	},
	Verbosity: logrus.InfoLevel.String(),
}

func Parse(r io.Reader) (Config, error) {
	cfg := defaultConfig
	return cfg, yaml.NewDecoder(r).Decode(&cfg)
}
