package main

import (
	"fmt"
	"os"

	"github.com/jessfraz/dash/googleanalytics"
	"github.com/sirupsen/logrus"
)

func doGoogleAnalytics() {
	if _, err := os.Stat(googleAnalyticsKeyfile); os.IsNotExist(err) {
		logrus.Warnf("Google Analytics keyfile %q does not exist", googleAnalyticsKeyfile)
		logrus.Info("skipping Google Analytics data")
		return
	}

	// Check that the Google Analytics view ID is not empty.
	if len(googleAnalyticsViewIDs) <= 0 {
		logrus.Warn("Google Analytics view ID cannot be empty")
		logrus.Info("skipping Google Analytics data")
		return
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
		fmt.Println()
	}
}
