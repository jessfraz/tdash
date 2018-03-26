package main

import (
	"fmt"
	"os"

	"github.com/gizak/termui"
	"github.com/jessfraz/tdash/googleanalytics"
	"github.com/sirupsen/logrus"
)

type gaData struct {
	name        string
	table       *termui.Table
	activeUsers string
}

func doGoogleAnalytics() ([]gaData, error) {
	if _, err := os.Stat(googleAnalyticsKeyfile); os.IsNotExist(err) {
		logrus.Warnf("Google Analytics keyfile %q does not exist", googleAnalyticsKeyfile)
		logrus.Info("skipping Google Analytics data")
		return nil, nil
	}

	// Check that the Google Analytics view ID is not empty.
	if len(googleAnalyticsViewIDs) <= 0 {
		logrus.Warn("Google Analytics view ID cannot be empty")
		logrus.Info("skipping Google Analytics data")
		return nil, nil
	}

	// Create the Google Analytics Client
	gaClient, err := googleanalytics.New(googleAnalyticsKeyfile, debug)
	if err != nil {
		return nil, fmt.Errorf("creating Google Analytics client failed: %v", err)
	}

	// Iterate over the Google Analytics view IDs.
	data := []gaData{}
	for _, gaViewID := range googleAnalyticsViewIDs {
		// Initialize our gaData.
		ga := gaData{}

		// Get the name of our Google Analytics view ID.
		ga.name, err = gaClient.GetProfileName(gaViewID)
		if err != nil {
			return nil, fmt.Errorf("getting Google Analytics view name for %q failed: %v", gaViewID, err)
		}

		// Get the Google Analytics report.
		resp, err := gaClient.GetReport(gaViewID)
		if err != nil {
			return nil, fmt.Errorf("getting Google Analytics report for view %q failed: %v", gaViewID, err)
		}

		// Create a termui Widget from the Google Analytics report.
		// TODO(jessfraz): make setting the max rows a flag.
		ga.table, err = googleanalytics.CreateWidget(resp, 10)
		if err != nil {
			return nil, fmt.Errorf("printing Google Analytics response failed: %v", err)
		}

		// Get the realtime data for users.
		ga.activeUsers, err = gaClient.GetRealtimeActiveUsers(gaViewID)
		if err != nil {
			return nil, fmt.Errorf("getting Google Analytics realtime active users data for view %q failed: %v", gaViewID, err)
		}

		// Append to our data.
		data = append(data, ga)
	}

	return data, nil
}

func gaWidget(body *termui.Grid) {
	if body == nil {
		body = termui.Body
	}

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
			body.AddRows(
				termui.NewRow(termui.NewCol(9, 0, data.table), termui.NewCol(3, 0, activeUsers)),
			)
		}
	}
	// Calculate the layout.
	body.Align()
	// Render the termui body.
	termui.Render(body)
}
