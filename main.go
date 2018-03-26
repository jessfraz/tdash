package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gizak/termui"
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
	flag.BoolVar(&showAllBuilds, "all", false, "Show all builds even successful ones, defaults to only showing failures")

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
	// Initialize termui.
	if err := termui.Init(); err != nil {
		logrus.Fatalf("initializing termui failed: %v", err)
	}
	defer termui.Close()

	// Create termui widgets for google analytics.
	go func() {
		ga, err := doGoogleAnalytics()
		if err != nil {
			logrus.Fatal(err)
		}

		// Add Google Analytics data to the termui body.
		for _, data := range ga {
			data.table.Block.BorderLabel = "Google Analytics data for " + data.name

			activeUsers := termui.NewPar(data.activeUsers)
			activeUsers.TextFgColor = termui.ColorWhite
			activeUsers.BorderFg = termui.ColorWhite
			activeUsers.BorderLabel = "active users for " + data.name
			activeUsers.Height = 3
			activeUsers.Width = 50

			if data.table != nil {
				termui.Body.AddRows(
					termui.NewRow(termui.NewCol(9, 0, data.table), termui.NewCol(3, 0, activeUsers)),
				)
			}
		}
		// Calculate the layout.
		termui.Body.Align()
		// Render the termui body.
		termui.Render(termui.Body)
	}()

	// Create termui widgets for travis.
	go func() {
		travis, err := doTravisCI()
		if err != nil {
			logrus.Fatal(err)
		}
		if travis != nil {
			columns := []*termui.Row{}
			for _, t := range travis {
				columns = append(columns, termui.NewCol(12/len(travis), 0, t))
			}
			termui.Body.AddRows(termui.NewRow(columns...))

			// Calculate the layout.
			termui.Body.Align()
			// Render the termui body.
			termui.Render(termui.Body)
		}
	}()

	go func() {
		janky, err := doJenkinsCI()
		if err != nil {
			logrus.Fatal(err)
		}
		if janky != nil {
			termui.Body.AddRows(termui.NewCol(3, 0, janky))

			// Calculate the layout.
			termui.Body.Align()
			// Render the termui body.
			termui.Render(termui.Body)
		}
	}()

	// Calculate the layout.
	termui.Body.Align()
	// Render the termui body.
	termui.Render(termui.Body)

	// Handle key q pressing
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		// press q to quit
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		// handle Ctrl + c combination
		termui.StopLoop()
	})

	// Handle resize
	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})

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
