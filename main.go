package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jessfraz/dash/googleanalytics"
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
	googleanalyticsViewID  string

	dashDir string

	debug bool
	vrsn  bool
)

func init() {
	// Get home directory.
	home, err := getHome()
	if err != nil {
		logrus.Fatal(err)
	}
	dashDir = filepath.Join(home, ".dash")

	// Parse flags.
	flag.StringVar(&googleAnalyticsKeyfile, "ga-keyfile", filepath.Join(dashDir, "ga.json"), "Path to Google Analytics keyfile")
	flag.StringVar(&googleanalyticsViewID, "ga-viewid", os.Getenv("GOOGLE_ANALYTICS_VIEW_ID"), "Google Analytics view ID (or env var GOOGLE_ANALYTICS_VIEW_ID)")

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
	if _, err := os.Stat(googleAnalyticsKeyfile); os.IsNotExist(err) {
		// TODO(jessfraz): make this just not get google analytics data
		// if we don't have a file and just warn the user.
		logrus.Fatal("Google Analytics keyfile %q does not exist", googleAnalyticsKeyfile)
	}

	// Check that the Google Analytics view ID is not empty.
	if googleanalyticsViewID == "" {
		logrus.Fatal("Google Analytics view ID cannot be empty")
	}

	// Create the Google Analytics Client
	gaClient, err := googleanalytics.New(googleAnalyticsKeyfile, debug)
	if err != nil {
		logrus.Fatalf("creating Google Analytics client failed: %v", err)
	}

	// Get the Google Analytics report.
	resp, err := gaClient.GetReport(googleanalyticsViewID)
	if err != nil {
		logrus.Fatalf("getting Google Analytics report for view %q failed: %v", googleanalyticsViewID, err)
	}

	// Print the Google Analytics report.
	// TODO(jessfraz): make setting the max rows a flag.
	if err := googleanalytics.PrintResponse(resp, 20); err != nil {
		logrus.Fatalf("printing Google Analytics response failed: %v", err)
	}
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
