## go-circleci
[![GoDoc](https://godoc.org/github.com/jszwedko/go-circleci?status.svg)](http://godoc.org/github.com/jszwedko/go-circleci)
[![Circle CI](https://circleci.com/gh/jszwedko/go-circleci.svg?style=svg)](https://circleci.com/gh/jszwedko/go-circleci)
[![Go Report Card](https://goreportcard.com/badge/github.com/jszwedko/go-circleci)](https://goreportcard.com/report/github.com/jszwedko/go-circleci)
[![coverage](https://gocover.io/_badge/github.com/jszwedko/go-circleci?0 "coverage")](http://gocover.io/github.com/jszwedko/go-circleci)

Go library for interacting with [CircleCI's API](https://circleci.com/docs/api). Supports all current API endpoints allowing you do do things like:

* Query for recent builds
* Get build details
* Retry builds
* Manipulate checkout keys, environment variables, and other settings for a project

**The CircleCI HTTP API response schemas are not well documented so please file an issue if you run into something that doesn't match up.**

Example usage:

```golang
package main

import (
        "fmt"

        "github.com/jszwedko/go-circleci"
)

func main() {
        client := &circleci.Client{Token: "YOUR TOKEN"} // Token not required to query info for public projects

        builds, _ := client.ListRecentBuildsForProject("jszwedko", "circleci-cli", "master", "", -1, 0)

        for _, build := range builds {
                fmt.Printf("%d: %s\n", build.BuildNum, build.Status)
        }
}
```

For the CLI that uses this library (or to see more example usages), please see
[circleci-cli](https://github.com/jszwedko/circleci-cli).

Currently in alpha, so the library API may change -- please use your favorite
Go dependency management solution.

See [GoDoc](http://godoc.org/github.com/jszwedko/go-circleci) for API usage.

Feature requests and issues welcome!
