package main

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/nlowe/pianoman/cmd"
)

//go:generate mockery

func main() {
	logrus.SetFormatter(&prefixed.TextFormatter{
		FullTimestamp:   true,
		ForceFormatting: true,
	})
	logrus.SetOutput(os.Stderr)

	if cmd.NewRootCmd().Execute() != nil {
		os.Exit(1)
	}
}
