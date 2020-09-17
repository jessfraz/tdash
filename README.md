# tdash

[![make-all](https://github.com/jessfraz/tdash/workflows/make%20all/badge.svg)](https://github.com/jessfraz/tdash/actions?query=workflow%3A%22make+all%22)
[![make-image](https://github.com/jessfraz/tdash/workflows/make%20image/badge.svg)](https://github.com/jessfraz/tdash/actions?query=workflow%3A%22make+image%22)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/tdash)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/tdash/total.svg?style=for-the-badge)](https://github.com/jessfraz/tdash/releases)

A terminal dashboard with stats from Google Analytics, GitHub, Travis CI, and Jenkins. Very much built specific to me.

![term.png](term.png)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Installation](#installation)
    - [Binaries](#binaries)
    - [Via Go](#via-go)
    - [Running with Docker](#running-with-docker)
- [Usage](#usage)
- [Setup](#setup)
  - [Google Analytics](#google-analytics)
  - [Travis](#travis)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/tdash/releases).

#### Via Go

```console
$ go get github.com/jessfraz/tdash
```

#### Running with Docker

```console
$ docker run --rm -it \
    --name tdash \
    --volume /home/user/.tdash:/root/.tdash:ro \ # for the  Google Analytics key file
    r.j3ss.co/tdash
```

## Usage

```conosle
$ tdash -h
tdash -  A terminal dashboard with stats from Google Analytics, GitHub, Travis CI, and Jenkins.

Usage: tdash <command>

Flags:

  --travis-owner      Travis owner name for builds (can have more than one) (default: [])
  -d                  enable debug logging (default: false)
  --ga-viewid         Google Analytics view IDs (can have more than one) (default: [])
  --interval          update interval (ex. 5ms, 10s, 1m, 3h) (default: 2m0s)
  --jenkins-password  Jenkins password for authentication (or env var JENKINS_PASSWORD)
  --jenkins-uri       Jenkins base URI (or env var JENKINS_BASE_URI)
  --jenkins-username  Jenkins username for authentication (or env var JENKINS_USERNAME)
  --all               Show all builds even successful ones, defaults to only showing failures (default: false)
  --ga-keyfile        Path to Google Analytics keyfile (default: ~/.tdash/ga.json)
  --travis-token      Travis CI API token (or env var TRAVISCI_API_TOKEN)

Commands:

  version  Show the version information.
```

## Setup

### Google Analytics

1. Enable the API: To get started using Analytics Reporting API v4, you need to 
    first create a project in the 
    [Google API Console](https://console.developers.google.com),
    enable the API, and create credentials.

    Follow the instructions 
    [for step enabling the API here](https://developers.google.com/anaytics/devguides/reporting/core/v4/quickstart/service-java).

2. Add the new service account to the Google Analytics account with 
    [Read & Analyze](https://support.google.com/analytics/answer/2884495) 
    permission.

    The newly created service account will have an email address that looks
    similar to: `quickstart@PROJECT-ID.iam.gserviceaccount.com`.

    Use this email address to 
    [add a user](https://support.google.com/analytics/answer/1009702) to the 
    Google Analytics view you want to access via the API. 

### Travis

1. Get your Travis token: Go to the "Profile" tab on your 
	[Accounts page](https://travis-ci.org/profile)
