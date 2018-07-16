package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/genuinetools/pkg/cli"
	"github.com/gizak/termui"
	"github.com/jessfraz/tdash/version"
	"github.com/sirupsen/logrus"
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
	interval      time.Duration

	dashDir string

	debug bool
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

func main() {
	// Get home directory.
	home, err := getHome()
	if err != nil {
		logrus.Fatal(err)
	}
	dashDir = filepath.Join(home, ".tdash")

	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "tdash"
	p.Description = " A terminal dashboard with stats from Google Analytics, GitHub, Travis CI, and Jenkins"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.BoolVar(&showAllBuilds, "all", false, "Show all builds even successful ones, defaults to only showing failures")
	p.FlagSet.DurationVar(&interval, "interval", 2*time.Minute, "update interval (ex. 5ms, 10s, 1m, 3h)")

	p.FlagSet.StringVar(&googleAnalyticsKeyfile, "ga-keyfile", filepath.Join(dashDir, "ga.json"), "Path to Google Analytics keyfile")
	p.FlagSet.Var(&googleAnalyticsViewIDs, "ga-viewid", "Google Analytics view IDs (can have more than one)")

	p.FlagSet.StringVar(&travisToken, "travis-token", os.Getenv("TRAVISCI_API_TOKEN"), "Travis CI API token (or env var TRAVISCI_API_TOKEN)")
	p.FlagSet.Var(&travisOwners, "travis-owner", "Travis owner name for builds (can have more than one)")

	p.FlagSet.StringVar(&jenkinsBaseURI, "jenkins-uri", os.Getenv("JENKINS_BASE_URI"), "Jenkins base URI (or env var JENKINS_BASE_URI)")
	p.FlagSet.StringVar(&jenkinsUsername, "jenkins-username", os.Getenv("JENKINS_USERNAME"), "Jenkins username for authentication (or env var JENKINS_USERNAME)")
	p.FlagSet.StringVar(&jenkinsPassword, "jenkins-password", os.Getenv("JENKINS_PASSWORD"), "Jenkins password for authentication (or env var JENKINS_PASSWORD)")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		ticker := time.NewTicker(interval)

		// Initialize termui.
		if err := termui.Init(); err != nil {
			logrus.Fatalf("initializing termui failed: %v", err)
		}
		defer termui.Close()

		go doWidgets()

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
			doWidgets()
		})

		// Update on an interval
		go func() {
			for range ticker.C {
				doWidgets()
			}
		}()

		// Start the loop.
		termui.Loop()
		return nil
	}

	// Run our program.
	p.Run()
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

func doWidgets() {
	body := termui.NewGrid()
	body.X = 0
	body.Y = 0
	body.BgColor = termui.ThemeAttr("bg")
	body.Width = termui.TermWidth()

	ga, err := doGoogleAnalytics()
	if err != nil {
		termui.StopLoop()
		termui.Close()
		logrus.Fatal(err)
	}

	// Add Google Analytics data to the termui body.
	for _, data := range ga {
		data.table.Block.BorderLabel = "Google Analytics data for " + data.name

		activeUsers := termui.NewPar(data.activeUsers)
		activeUsers.TextFgColor = termui.ColorWhite
		activeUsers.BorderFg = termui.ColorWhite
		activeUsers.BorderLabel = "Active users for " + data.name
		activeUsers.Height = 3

		if data.table != nil {
			body.AddRows(
				termui.NewRow(termui.NewCol(9, 0, data.table), termui.NewCol(3, 0, activeUsers)),
			)
		}
	}

	travis, err := doTravisCI()
	if err != nil {
		termui.StopLoop()
		termui.Close()
		logrus.Fatal(err)
	}
	if travis != nil {
		columns := []*termui.Row{}
		for _, t := range travis {
			columns = append(columns, termui.NewCol(int(12/len(travis)), 0, t))
		}
		body.AddRows(termui.NewRow(columns...))
	}

	janky, err := doJenkinsCI()
	if err != nil {
		termui.StopLoop()
		termui.Close()
		logrus.Fatal(err)
	}
	if janky != nil {
		body.AddRows(termui.NewCol(3, 0, janky))
	}

	// Calculate the layout.
	body.Align()
	// Render the termui body.
	termui.Clear()
	termui.Render(body)
}
