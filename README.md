# dash

[![Travis CI](https://travis-ci.org/jessfraz/dash.svg?branch=master)](https://travis-ci.org/jessfraz/dash)

A terminal dashboard with stats from Google Analytics, GitHub, Travis CI, and Jenkins. Very much built specific to me.

## Installation

#### Binaries

- **darwin** [386](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-darwin-386) / [amd64](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-darwin-amd64)
- **freebsd** [386](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-freebsd-386) / [amd64](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-freebsd-amd64)
- **linux** [386](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-linux-386) / [amd64](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-linux-amd64) / [arm](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-linux-arm) / [arm64](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-linux-arm64)
- **solaris** [amd64](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-solaris-amd64)
- **windows** [386](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-windows-386) / [amd64](https://github.com/jessfraz/dash/releases/download/v0.0.0/dash-windows-amd64)

#### Via Go

```bash
$ go get github.com/jessfraz/dash
```

#### Running with Docker

```console
$ docker run --rm -it \
    --name dash \
    --volume /home/user/.dash:/root/.dash:ro \ # for the  Google Analytics key file
    r.j3ss.co/dash
```

## Usage

```conosle
$ dash -h
     _           _
  __| | __ _ ___| |__
 / _` |/ _` / __| '_ \
| (_| | (_| \__ \ | | |
 \__,_|\__,_|___/_| |_|


 A single page dashboard with stats from
 Google Analytics, GitHub, Travis CI, and Jenkins.
 Version: v0.0.0
 Build: 338723b

  -all
        Show all builds even successful ones, defaults to only showing failures
  -d    run in debug mode
  -ga-keyfile string
        Path to Google Analytics keyfile (default "/home/jessie/.dash/ga.json")
  -ga-viewid value
        Google Analytics view IDs (can have more than one)
  -jenkins-password string
        Jenkins password for authentication (or env var JENKINS_PASSWORD)
  -jenkins-uri string
        Jenkins base URI (or env var JENKINS_BASE_URI)
  -jenkins-username string
        Jenkins username for authentication (or env var JENKINS_USERNAME)
  -travis-owner value
        Travis owner name for builds (can have more than one)
  -travis-token string
        Travis CI API token (or env var TRAVISCI_API_TOKEN)
  -v    print version and exit (shorthand)
  -version
        print version and exit
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
