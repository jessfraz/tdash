package main

import (
	"fmt"

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

	// Iterate over the travisOwners if it was passed.
	for _, travisOwner := range travisOwners {
		travisClient := travis.NewClient(travis.TRAVIS_API_DEFAULT_URL, "")
		repos, _, err := travisClient.Repositories.Find(&travis.RepositoryListOptions{
			OwnerName: travisOwner,
		})
		if err != nil {
			logrus.Fatalf("getting travis repos for owner %q failed: %v", travisOwner, err)
		}

		fmt.Printf("repos: %#v\n", repos)
	}
}
