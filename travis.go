package main

import (
	"context"
	"fmt"
	"net/http"

	travis "github.com/Ableton/go-travis"
	"github.com/gizak/termui"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

func doTravisCI() ([]*termui.Table, error) {
	// Check that the Travis CI API token is not empty.
	if len(travisToken) <= 0 {
		logrus.Warn("Travis CI API token cannot be empty")
		logrus.Info("skipping Travis CI data")
		return nil, nil
	}

	// Check that the Travis owners is not empty.
	if len(travisOwners) <= 0 {
		logrus.Warn("Travis CI owners cannot be empty")
		logrus.Info("skipping Travis CI data")
		return nil, nil
	}

	tables := []*termui.Table{}

	// Iterate over the travisOwners if it was passed.
	for _, travisOwner := range travisOwners {
		// Initialize the table.
		table := termui.NewTable()
		rows := [][]string{
			{"repo", "branch", "state", "finished at"},
		}
		badrows := []int{}

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
				return nil, fmt.Errorf("listing repos for %q failed: %v", travisOwner, err)
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
				return nil, fmt.Errorf("getting master branch for travis repo %q failed: %v", repo.GetFullName(), err)
			}

			rows = append(rows, []string{repo.GetFullName(), "master", branch.State, branch.FinishedAt})

			if branch.State != "passed" {
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
		table.Separator = true
		table.Block.BorderLabel = "Travis CI builds for " + travisOwner
		table.Analysis()
		table.SetSize()
		// Set the color to red for the bad rows
		for _, br := range badrows {
			table.FgColors[br] = termui.ColorRed
		}

		tables = append(tables, table)
	}

	return tables, nil
}
