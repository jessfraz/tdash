package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jessfraz/dash/version"
	"github.com/sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `     _           _
  __| | __ _ ___| |__
 / _` + "`" + ` |/ _` + "`" + ` / __| '_ \
| (_| | (_| \__ \ | | |
 \__,_|\__,_|___/_| |_|


 A single page dashboard with stats from
 Google Analytics, GitHub, Travis CI, and Jenkins.
 Version: %s
 Build: %s

`
)

var (
	debug bool
	vrsn  bool
)

func init() {
	// parse flags
	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION, version.GITCOMMIT))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vrsn {
		fmt.Printf("dash version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
}
