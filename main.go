package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

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
	googleAnalyticsKeyfile string
	googleAnalyticsViewIDs stringSlice

	travisToken  string
	travisOwners stringSlice

	jenkinsBaseURI  string
	jenkinsUsername string
	jenkinsPassword string

	dashDir string

	debug bool
	vrsn  bool
)

// stringSlice is a slice of strings
type stringSlice []string

// implement the flag interface for stringSlice
func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func init() {
	// Get home directory.
	home, err := getHome()
	if err != nil {
		logrus.Fatal(err)
	}
	dashDir = filepath.Join(home, ".dash")

	// Parse flags.
	flag.StringVar(&googleAnalyticsKeyfile, "ga-keyfile", filepath.Join(dashDir, "ga.json"), "Path to Google Analytics keyfile")
	flag.Var(&googleAnalyticsViewIDs, "ga-viewid", "Google Analytics view IDs (can have more than one)")

	flag.StringVar(&travisToken, "travis-token", os.Getenv("TRAVISCI_API_TOKEN"), "Travis CI API token (or env var TRAVISCI_API_TOKEN)")
	flag.Var(&travisOwners, "travis-owner", "Travis owner name for builds (can have more than one)")

	flag.StringVar(&jenkinsBaseURI, "jenkins-uri", os.Getenv("JENKINS_BASE_URI"), "Jenkins base URI (or env var JENKINS_BASE_URI)")
	flag.StringVar(&jenkinsUsername, "jenkins-username", os.Getenv("JENKINS_USERNAME"), "Jenkins username for authentication (or env var JENKINS_USERNAME)")
	flag.StringVar(&jenkinsPassword, "jenkins-password", os.Getenv("JENKINS_PASSWORD"), "Jenkins password for authentication (or env var JENKINS_PASSWORD)")

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

	// Set the log level.
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	doGoogleAnalytics()
	doTravisCI()
	doJenkinsCI()
}

func getHome() (string, error) {
	home := os.Getenv(homeKey)
	if home != "" {
		return home, nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}
