package main

import (
	"fmt"

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

	// Get info about jenkins
	info, err := jenkinsClient.Info()
	if err != nil {
		logrus.Fatalf("getting all jenkins info failed: %v", err)
	}

	fmt.Printf("info: %#v\n", info)

	// Get all the jobs
	jobs, err := jenkinsClient.GetAllJobs()
	if err != nil {
		logrus.Fatalf("getting all jenkins jobs failed: %v", err)
	}

	fmt.Printf("jobs: %#v\n", jobs)
}
