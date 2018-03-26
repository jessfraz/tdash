package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	travis "github.com/Ableton/go-travis"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

func doTravisCI() {
	// Check that the Travis CI API token is not empty.
	if len(travisToken) <= 0 {
		logrus.Warn("Travis CI API token cannot be empty")
		logrus.Info("skipping Travis CI data")
		return
	}

	// Check that the Travis owners is not empty.
	if len(travisOwners) <= 0 {
		logrus.Warn("Travis CI owners cannot be empty")
		logrus.Info("skipping Travis CI data")
		return
	}

	// Create the tabwriter.
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

	// Print dimensions and metrics header.
	fmt.Fprintln(w, "REPO\tBRANCH\tSTATE\tFINISHED AT")

	// Iterate over the travisOwners if it was passed.
	for _, travisOwner := range travisOwners {
		// Get the owners repos from GitHub.
		ghClient := github.NewClient(nil)
		opt := &github.RepositoryListOptions{
			ListOptions: github.ListOptions{PerPage: 100},
			Type:        "sources",
		}
		var repos []*github.Repository
		for {
			reposResp, resp, err := ghClient.Repositories.List(context.Background(), travisOwner, opt)
			if err != nil {
				logrus.Fatalf("listing repos for %q failed: %v", travisOwner, err)
			}
			repos = append(repos, reposResp...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		// Initialize the travis client.
		travisClient := travis.NewClient(travis.TRAVIS_API_DEFAULT_URL, travisToken)

		// Iterate over the repositories and get the master branch build status.
		for _, repo := range repos {
			if repo.GetFork() {
				// Continue early if its a fork because we don't care
				continue
			}

			// Get the master branch
			branch, resp, err := travisClient.Branches.GetFromSlug(repo.GetFullName(), "master")
			if err != nil {
				// This will fail on forks or non travis building repos with a 404
				// so we might as well error silently if we get a 404.
				if resp.StatusCode == http.StatusNotFound {
					continue
				}
				logrus.Fatalf("getting master branch for travis repo %q failed: %v", repo.GetFullName(), err)
			}

			if showAllBuilds || branch.State != "passed" {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repo.GetFullName(), "master", branch.State, branch.FinishedAt)
			}
		}
	}

	w.Flush()
	fmt.Println()
}
