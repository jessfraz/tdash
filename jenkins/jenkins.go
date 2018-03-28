package jenkins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Client contains the information for connecting to a jenkins instance
type Client struct {
	Baseurl  string `json:"base_url"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// JobsResponse describes a response for jobs.
type JobsResponse struct {
	Jobs []Job `json:"jobs,omitempty"`
}

// Job describes a job object from the Jenkins API.
type Job struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	LastBuild   Build  `json:"lastBuild,omitempty"`
}

// Build describes a build from the Jenkins API.
type Build struct {
	Result    string `json:"result,omitempty"`
	Number    int    `json:"number,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// New sets the authentication for the Jenkins client
// Password can be an API token as described in:
// https://wiki.jenkins-ci.org/display/JENKINS/Authenticating+scripted+clients
func New(uri, username, token string) *Client {
	return &Client{
		Baseurl:  uri,
		Username: username,
		Token:    token,
	}
}

// GetJobs gets the jobs for a Jenkins instance.
func (c *Client) GetJobs() ([]Job, error) {
	// set up the request
	url := fmt.Sprintf("%s/api/json?tree=%s&depth=1", c.Baseurl, url.QueryEscape("jobs[name,displayName,lastBuild[number,timestamp,result]]"))
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	// add the auth
	req.SetBasicAuth(c.Username, c.Token)

	// do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// check the status code
	// it should be 200
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("jenkins get jobs request to %s responded with status %d", url, resp.StatusCode)
	}

	var r JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decoding json response from jobs from %s failed: %v", url, err)
	}

	return r.Jobs, nil
}
