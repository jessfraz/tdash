package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/gizak/termui"
	"github.com/jessfraz/tdash/version"
	"github.com/sirupsen/logrus"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `     _           _
  __| | __ _ ___| |__
 / _` + "`" + ` |/ _` + "`" + ` / __| '_ \
| (_| | (_| \__ \ | | |
 \__,_|\__,_|___/_| |_|


 A terminal dashboard with stats from
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

	showAllBuilds bool
	interval      string

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
	dashDir = filepath.Join(home, ".tdash")

	// Parse flags.
	flag.BoolVar(&showAllBuilds, "all", false, "Show all builds even successful ones, defaults to only showing failures")
	flag.StringVar(&interval, "interval", "2m", "update interval (ex. 5ms, 10s, 1m, 3h)")

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
		fmt.Printf("tdash version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	// Set the log level.
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	var ticker *time.Ticker

	// parse the duration
	dur, err := time.ParseDuration(interval)
	if err != nil {
		logrus.Fatalf("parsing %s as duration failed: %v", interval, err)
	}
	ticker = time.NewTicker(dur)

	// Initialize termui.
	if err := termui.Init(); err != nil {
		logrus.Fatalf("initializing termui failed: %v", err)
	}
	defer termui.Close()

	// Create termui widgets for google analytics.
	go gaWidget(nil)
	go travisWidget(nil)
	go jenkinsWidget(nil)

	// Calculate the layout.
	termui.Body.Align()
	// Render the termui body.
	termui.Render(termui.Body)

	// Handle key q pressing
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		// press q to quit
		ticker.Stop()
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		// handle Ctrl + c combination
		ticker.Stop()
		termui.StopLoop()
	})

	// Handle resize
	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})

	// Update on an interval
	go func() {
		for range ticker.C {
			body := termui.NewGrid()
			body.X = 0
			body.Y = 0
			body.BgColor = termui.ThemeAttr("bg")
			body.Width = termui.TermWidth()

			gaWidget(body)
			travisWidget(body)
			jenkinsWidget(body)

			// Calculate the layout.
			body.Align()
			// Render the termui body.
			termui.Render(body)
		}
	}()

	// Start the loop.
	termui.Loop()
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
