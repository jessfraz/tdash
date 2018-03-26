package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	travis "github.com/Ableton/go-travis"
	"github.com/sirupsen/logrus"
)

func doTravisCI() {
	// Check that the Travis CI API token is not empty.
	if len(travisToken) <= 0 {
		logrus.Warn("Travis CI API token cannot be empty")
		logrus.Info("skipping Travis CI data")
		return
	}

	// Create the tabwriter.
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

	// Print dimensions and metrics header.
	fmt.Fprintln(w, "REPO\tBRANCH\tSTATE\tFINISHED AT")

	// Iterate over the travisOwners if it was passed.
	for _, travisOwner := range travisOwners {
		travisClient := travis.NewClient(travis.TRAVIS_API_DEFAULT_URL, travisToken)
		repos, _, err := travisClient.Repositories.Find(&travis.RepositoryListOptions{
			OwnerName: travisOwner,
			Active:    true,
		})
		if err != nil {
			logrus.Fatalf("getting travis repos for owner %q failed: %v", travisOwner, err)
		}

		// Iterate over the repositories and get the master branch build status.
		for _, repo := range repos {
			// Get the master branch
			branch, _, err := travisClient.Branches.GetFromSlug(repo.Slug, "master")
			if err != nil {
				// This will fail on forks with a 404 so we might as well error silently
				// unless in debug mode.
				logrus.Debugf("getting master branch for travis repo %q failed: %v", repo.Slug, err)
				continue
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repo.Slug, "master", branch.State, branch.FinishedAt)
		}
	}

	w.Flush()
}
