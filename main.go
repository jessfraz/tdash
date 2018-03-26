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
	googleAnalyticsViewIDs stringSlice

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
	if len(googleAnalyticsViewIDs) <= 0 {
		logrus.Fatal("Google Analytics view ID cannot be empty")
	}

	// Create the Google Analytics Client
	gaClient, err := googleanalytics.New(googleAnalyticsKeyfile, debug)
	if err != nil {
		logrus.Fatalf("creating Google Analytics client failed: %v", err)
	}

	// Iterate over the Google Analytics view IDs.
	for _, gaViewID := range googleAnalyticsViewIDs {
		// Get the name of our Google Analytics view ID.
		gaViewName, err := gaClient.GetProfileName(gaViewID)
		if err != nil {
			logrus.Fatalf("getting Google Analytics view name for %q failed: %v", gaViewID, err)
		}

		fmt.Printf("Google Analytics data for view %s\n\n", gaViewName)

		// Get the Google Analytics report.
		resp, err := gaClient.GetReport(gaViewID)
		if err != nil {
			logrus.Fatalf("getting Google Analytics report for view %q failed: %v", gaViewID, err)
		}

		// Print the Google Analytics report.
		// TODO(jessfraz): make setting the max rows a flag.
		if err := googleanalytics.PrintResponse(resp, 20); err != nil {
			logrus.Fatalf("printing Google Analytics response failed: %v", err)
		}

		// Get the realtime data for users.
		activeUsers, err := gaClient.GetRealtimeActiveUsers(gaViewID)
		if err != nil {
			logrus.Fatalf("getting Google Analytics realtime active users data for view %q failed: %v", gaViewID, err)
		}

		fmt.Printf("\nRealtime Active Users: %s\n", activeUsers)
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
