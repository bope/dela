package main

import (
	"os"

	"github.com/bope/dela/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var shareDir string
var mountDir string
var verbosity int

func init() {
	pflag.StringVarP(&shareDir, "share", "s", "", "share directory")
	pflag.StringVarP(&mountDir, "mount", "m", "", "fuse mount directory")
	pflag.CountVarP(&verbosity, "verbose", "v", "logging verbosity")
}

func main() {
	pflag.Parse()

	switch verbosity {
	case 0:
		log.SetLevel(log.ErrorLevel)
	case 1:
		log.SetLevel(log.InfoLevel)
	case 2:
		log.SetLevel(log.DebugLevel)
	}

	if shareDir == "" && mountDir == "" {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	if err := cmd.Run(mountDir, shareDir); err != nil {
		os.Exit(1)
	}
}
