package main

import (
	"fmt"
	"time"

	jenkins "github.com/bndr/gojenkins"
	"github.com/gizak/termui"
	"github.com/sirupsen/logrus"
)

func doJenkinsCI() (*termui.Table, error) {
	// Check that the jenkins base URI is not empty.
	if len(jenkinsBaseURI) <= 0 {
		logrus.Warn("Jenkins Base URI cannot be empty")
		logrus.Info("skipping Jenkins CI data")
		return nil, nil
	}

	// Check that the jenkins username is not empty.
	if len(jenkinsUsername) <= 0 {
		logrus.Warn("Jenkins username cannot be empty")
		logrus.Info("skipping Jenkins CI data")
		return nil, nil
	}

	// Check that the jenkins password is not empty.
	if len(jenkinsPassword) <= 0 {
		logrus.Warn("Jenkins password cannot be empty")
		logrus.Info("skipping Jenkins CI data")
		return nil, nil
	}

	// Initialize the jenkins api client
	jenkinsClient, err := jenkins.CreateJenkins(nil, jenkinsBaseURI, jenkinsUsername, jenkinsPassword).Init()
	if err != nil {
		return nil, fmt.Errorf("creating jenkins api client for base uri %q failed: %v", jenkinsBaseURI, err)
	}

	// Get all the jobs
	jobs, err := jenkinsClient.GetAllJobs()
	if err != nil {
		return nil, fmt.Errorf("getting all jenkins jobs failed: %v", err)
	}

	// Initialize the table.
	table := termui.NewTable()
	rows := [][]string{
		{"job", "state", "finished at"},
	}
	badrows := []int{}

	// Iterate over the jobs.
	for _, job := range jobs {
		// Get the last build
		build, err := job.GetLastBuild()
		if err != nil {
			return nil, fmt.Errorf("getting jenkins build number %d for job %q failed: %v", job.Raw.LastBuild.Number, job.Raw.Name, err)
		}

		if build.Raw.Result == "" {
			// Then the job is currently running.
			build.Raw.Result = "RUNNING"
		}

		if showAllBuilds || build.Raw.Result != "SUCCESS" {
			rows = append(rows, []string{job.Raw.DisplayName, build.Raw.Result, time.Unix(0, int64(time.Millisecond)*build.Raw.Timestamp).Format(time.RFC3339)})
			badrows = append(badrows, len(rows)-1)
		}
	}

	// Set the rows.
	table.Rows = rows

	// Set the default colors and settings.
	table.FgColor = termui.ColorWhite
	table.BgColor = termui.ColorDefault
	table.TextAlign = termui.AlignLeft
	table.Border = true
	table.Block.BorderLabel = "Jenkins builds for " + jenkinsBaseURI
	table.Analysis()
	table.SetSize()
	// Set the color to red for the bad rows
	for _, br := range badrows {
		table.FgColors[br] = termui.ColorRed
	}

	return table, nil
}
