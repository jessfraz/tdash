package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	jenkins "github.com/bndr/gojenkins"
	"github.com/sirupsen/logrus"
)

func doJenkinsCI() {
	// Check that the jenkins base URI is not empty.
	if len(jenkinsBaseURI) <= 0 {
		logrus.Warn("Jenkins Base URI cannot be empty")
		logrus.Info("skipping Jenkins CI data")
		return
	}

	// Check that the jenkins username is not empty.
	if len(jenkinsUsername) <= 0 {
		logrus.Warn("Jenkins username cannot be empty")
		logrus.Info("skipping Jenkins CI data")
		return
	}

	// Check that the jenkins password is not empty.
	if len(jenkinsPassword) <= 0 {
		logrus.Warn("Jenkins password cannot be empty")
		logrus.Info("skipping Jenkins CI data")
		return
	}

	// Initialize the jenkins api client
	jenkinsClient, err := jenkins.CreateJenkins(nil, jenkinsBaseURI, jenkinsUsername, jenkinsPassword).Init()
	if err != nil {
		logrus.Fatalf("creating jenkins api client for base uri %q failed: %v", jenkinsBaseURI, err)
	}

	// Get all the jobs
	jobs, err := jenkinsClient.GetAllJobs()
	if err != nil {
		logrus.Fatalf("getting all jenkins jobs failed: %v", err)
	}

	// Create the tabwriter.
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

	// Print dimensions and metrics header.
	fmt.Fprintln(w, "JOB\tSTATE\tFINISHED AT")

	// Iterate over the jobs.
	for _, job := range jobs {
		// Get the last build
		build, err := job.GetLastBuild()
		if err != nil {
			logrus.Fatalf("getting jenkins build number %d for job %q failed: %v", job.Raw.LastBuild.Number, job.Raw.Name, err)
		}

		if build.Raw.Result == "" {
			// Then the job is currently running.
			build.Raw.Result = "RUNNING"
		}

		if showAllBuilds || build.Raw.Result != "SUCCESS" {
			fmt.Fprintf(w, "%s\t%s\t%s\n", job.Raw.DisplayName, build.Raw.Result, time.Unix(build.Raw.Timestamp, 0).Format(time.RFC3339))
		}

	}

	w.Flush()
}
