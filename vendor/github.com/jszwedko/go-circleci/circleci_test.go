package circleci

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the CircleCI client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	url, err := url.Parse(server.URL)
	if err != nil {
		panic(fmt.Sprintf("couldn't parse test server URL: %s", server.URL))
	}

	client = &Client{BaseURL: url}
}

func teardown() {
	defer server.Close()
}

func testBody(t *testing.T, r *http.Request, want string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading request body: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("request Body is %s, want %s", got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("request method: %v, want %v", got, want)
	}
}

func testAPIError(t *testing.T, err error, statusCode int, message string) {
	if err == nil {
		t.Errorf("expected APIError but got nil")
	}
	switch err := err.(type) {
	case *APIError:
		want := &APIError{HTTPStatusCode: statusCode, Message: message}
		if !reflect.DeepEqual(err, want) {
			t.Errorf("error was %+v, want %+v", err, want)
		}
	default:
		t.Errorf("expected APIError but got %T: %+v", err, err)
	}
}

func testQueryIncludes(t *testing.T, r *http.Request, key, value string) {
	got := r.URL.Query().Get(key)
	if got != value {
		t.Errorf("expected query to include: %s=%s, got %s=%s", key, value, key, got)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %s, want %s", header, got, want)
	}
}

func TestClient_request(t *testing.T) {
	setup()
	defer teardown()
	client.Token = "ABCD"
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		testHeader(t, r, "Content-Type", "application/json")
		testQueryIncludes(t, r, "circle-token", "ABCD")
		fmt.Fprint(w, `{"login": "jszwedko"}`)
	})

	err := client.request("GET", "/me", &User{}, nil, nil)
	if err != nil {
		t.Errorf(`Client.request("GET", "/me", &User{}, nil, nil) errored with %s`, err)
	}
}

func TestClient_request_withDebug(t *testing.T) {
	setup()
	defer teardown()
	buf := bytes.NewBuffer(nil)
	client.Token = "ABCD"
	client.Debug = true
	client.Logger = log.New(buf, "", 0)
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		testHeader(t, r, "Content-Type", "application/json")
		testQueryIncludes(t, r, "circle-token", "ABCD")
		fmt.Fprint(w, `{"login": "jszwedko"}`)
	})

	err := client.request("GET", "/me", &User{}, nil, nil)
	if err != nil {
		t.Errorf(`Client.request("GET", "/me", &User{}, nil, nil) errored with %s`, err)
	}

	output := buf.String()

	t.Logf("debug output:\n%s", output)
	if !strings.Contains(output, "request:") {
		t.Error(`expected "request:" to appear in debug output`)
	}
	if !strings.Contains(output, "HTTP/1.1") {
		t.Error(`expected http request to appear in debug output`)
	}

	if !strings.Contains(output, "response:") {
		t.Error(`expected "response:" to appear in debug output`)
	}
	if !strings.Contains(output, "HTTP/1.1 200 OK") {
		t.Error(`expected http request to appear in debug output`)
	}
}

func TestClient_request_unauthenticated(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message": "You must log in first"}`)
	})

	err := client.request("GET", "/me", &User{}, nil, nil)
	testAPIError(t, err, http.StatusUnauthorized, "You must log in first")
}

func TestClient_request_noErrorMessage(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := client.request("GET", "/me", &User{}, nil, nil)
	testAPIError(t, err, http.StatusInternalServerError, "")
}

func TestClient_Me(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"login": "jszwedko"}`)
	})

	me, err := client.Me()
	if err != nil {
		t.Errorf("Client.Me returned error: %v", err)
	}

	want := &User{Login: "jszwedko"}
	if !reflect.DeepEqual(me, want) {
		t.Errorf("Client.Me returned %+v, want %+v", me, want)
	}
}

func TestClient_ListProjects(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"reponame": "foo"}]`)
	})

	projects, err := client.ListProjects()
	if err != nil {
		t.Errorf("Client.ListProjects() returned error: %v", err)
	}

	want := []*Project{{Reponame: "foo"}}
	if !reflect.DeepEqual(projects, want) {
		t.Errorf("Client.ListProjects() returned %+v, want %+v", projects, want)
	}
}

func TestClient_ListProjects_parseFeatureFlagsRaw(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		[{
			"reponame": "foo",
			"feature_flags": {
				"not-in-feature-flags": true
			}
		}]
		`)
	})

	projects, err := client.ListProjects()
	if err != nil {
		t.Errorf("Client.ListProjects() returned error: %v", err)
	}

	if projects[0].FeatureFlags.Raw()["not-in-feature-flags"] != true {
		t.Errorf("expected Client.ListProjects()[not-in-feature-flags] to be true, was %+v", projects[0].FeatureFlags.Raw()["not-in-feature-flags"])
	}
}

func TestClient_EnableProject(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/org-name/repo-name/enable", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
	})

	err := client.EnableProject("org-name", "repo-name")
	if err != nil {
		t.Errorf("Client.EnableProject() returned error: %v", err)
	}
}

func TestClient_DisableProject(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/org-name/repo-name/enable", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	err := client.DisableProject("org-name", "repo-name")
	if err != nil {
		t.Errorf("Client.EnableProject() returned error: %v", err)
	}
}

func TestClient_FollowProject(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/org-name/repo-name/follow", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"reponame": "repo-name"}`)
	})

	project, err := client.FollowProject("org-name", "repo-name")
	if err != nil {
		t.Errorf("Client.FollowProject() returned error: %v", err)
	}

	want := &Project{Reponame: "repo-name"}
	if !reflect.DeepEqual(project, want) {
		t.Errorf("Client.FollowProject() returned %+v, want %+v", project, want)
	}
}

func TestClient_GetProject(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[
			{"username": "jszwedko", "reponame": "bar"},
			{"username": "joe", "reponame": "foo"},
			{"username": "jszwedko", "reponame": "foo"}
		]`)
	})

	project, err := client.GetProject("jszwedko", "foo")
	if err != nil {
		t.Errorf("Client.GetProject returned error: %v", err)
	}

	want := &Project{Username: "jszwedko", Reponame: "foo"}
	if !reflect.DeepEqual(project, want) {
		t.Errorf("Client.GetProject(%+v, %+v) returned %+v, want %+v", "jszwedko", "foo", project, want)
	}
}

func TestClient_GetProject_noMatching(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[
			{"username": "jszwedko", "reponame": "bar"}
		]`)
	})

	project, err := client.GetProject("jszwedko", "foo")
	if err != nil {
		t.Errorf("Client.GetProject returned error: %v", err)
	}

	if project != nil {
		t.Errorf("Client.GetProject(%+v, %+v) returned %+v, want %+v", "jszwedko", "foo", project, nil)
	}
}

func TestClient_GetProject_urlDecodeBranches(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		// using Fprintf instead Fprint becouse `go vet` complains about a possible intention to use a formatted string
		fmt.Fprintf(w, `[
			{"username": "jszwedko", "reponame": "bar", "branches": {"apiv1%%2E1": {}}}
		]`)
	})

	project, err := client.GetProject("jszwedko", "bar")
	if err != nil {
		t.Errorf("Client.GetProject returned error: %v", err)
	}

	_, ok := project.Branches["apiv1.1"]
	if !ok {
		t.Errorf("expected Client.GetProject(%+v, %+v) to return branches containing 'apiv1.1'  got %+v", "jszwedko", "foo", project.Branches)
	}
}

func TestClient_recentBuilds_multiPage(t *testing.T) {
	setup()
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/recent-builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(200)
		switch requestCount {
		case 0:
			testQueryIncludes(t, r, "offset", "0")
			testQueryIncludes(t, r, "limit", "100")
			fmt.Fprint(w, fmt.Sprintf("[%s]", strings.Trim(strings.Repeat(`{"build_num": 123},`, 100), ",")))
		case 1:
			testQueryIncludes(t, r, "offset", "100")
			testQueryIncludes(t, r, "limit", "99")
			fmt.Fprint(w, fmt.Sprintf("[%s]", strings.Trim(strings.Repeat(`{"build_num": 123},`, 99), ",")))
		default:
			t.Errorf("Client.ListRecentBuilds(%+v, %+v) made more than two requests to /recent-builds", 199, 0)
		}
		requestCount++
	})

	builds, err := client.recentBuilds("recent-builds", nil, 199, 0)
	if err != nil {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned error: %v", 199, 0, err)
	}

	if len(builds) != 199 {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned %+v results, want %+v", 199, 0, len(builds), 99)
	}
}

func TestClient_recentBuilds_multiPageExhausted(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/recent-builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testQueryIncludes(t, r, "offset", "0")
		testQueryIncludes(t, r, "limit", "100")
		fmt.Fprint(w, fmt.Sprintf("[%s]", strings.Trim(strings.Repeat(`{"build_num": 123},`, 50), ",")))
	})

	builds, err := client.recentBuilds("recent-builds", nil, 199, 0)
	if err != nil {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned error: %v", 199, 0, err)
	}

	if len(builds) != 50 {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned %+v results, want %+v", 199, 0, len(builds), 50)
	}
}

func TestClient_recentBuilds_multiPageNoLimit(t *testing.T) {
	setup()
	defer teardown()

	requestCount := 0
	mux.HandleFunc("/recent-builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(200)
		switch requestCount {
		case 0:
			testQueryIncludes(t, r, "offset", "0")
			testQueryIncludes(t, r, "limit", "100")
			fmt.Fprint(w, fmt.Sprintf("[%s]", strings.Trim(strings.Repeat(`{"build_num": 123},`, 100), ",")))
		case 1:
			testQueryIncludes(t, r, "offset", "100")
			testQueryIncludes(t, r, "limit", "100")
			fmt.Fprint(w, fmt.Sprintf("[%s]", strings.Trim(strings.Repeat(`{"build_num": 123},`, 99), ",")))
		default:
			t.Errorf("Client.ListRecentBuilds(%+v, %+v) made more than two requests to /recent-builds", -1, 0)
		}
		requestCount++
	})

	builds, err := client.recentBuilds("recent-builds", nil, -1, 0)
	if err != nil {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned error: %v", -1, 0, err)
	}

	if len(builds) != 199 {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned %+v results, want %+v", -1, 0, len(builds), 199)
	}
}

func TestClient_ListRecentBuilds(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/recent-builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testQueryIncludes(t, r, "offset", "2")
		testQueryIncludes(t, r, "limit", "10")
		fmt.Fprint(w, `[{"build_num": 123}, {"build_num": 124}]`)
	})

	builds, err := client.ListRecentBuilds(10, 2)
	if err != nil {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned error: %v", 10, 2, err)
	}

	want := []*Build{{BuildNum: 123}, {BuildNum: 124}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Client.ListRecentBuilds(%+v, %+v) returned %+v, want %+v", 10, 2, builds, want)
	}
}

func TestClient_ListRecentBuildsForProject(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/foo/bar/tree/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testQueryIncludes(t, r, "filter", "running")
		testQueryIncludes(t, r, "offset", "0")
		testQueryIncludes(t, r, "limit", "10")
		fmt.Fprint(w, `[{"build_num": 123}, {"build_num": 124}]`)
	})

	call := fmt.Sprintf("Client.ListRecentBuilds(foo, bar, master, running, 10, 0)")

	builds, err := client.ListRecentBuildsForProject("foo", "bar", "master", "running", 10, 0)
	if err != nil {
		t.Errorf("%s returned error: %v", call, err)
	}

	want := []*Build{{BuildNum: 123}, {BuildNum: 124}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("%s returned %+v, want %+v", call, builds, want)
	}
}

func TestClient_ListRecentBuildsForProject_noBranch(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/foo/bar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testQueryIncludes(t, r, "filter", "running")
		testQueryIncludes(t, r, "offset", "0")
		testQueryIncludes(t, r, "limit", "10")
		fmt.Fprint(w, `[{"build_num": 123}, {"build_num": 124}]`)
	})

	call := fmt.Sprintf("Client.ListRecentBuilds(foo, bar, , running, 10, 0)")

	builds, err := client.ListRecentBuildsForProject("foo", "bar", "", "running", 10, 0)
	if err != nil {
		t.Errorf("%s returned error: %v", call, err)
	}

	want := []*Build{{BuildNum: 123}, {BuildNum: 124}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("%s returned %+v, want %+v", call, builds, want)
	}
}

func TestClient_GetBuild(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"build_num": 123}`)
	})

	build, err := client.GetBuild("jszwedko", "foo", 123)
	if err != nil {
		t.Errorf("Client.GetBuild(jszwedko, foo, 123) returned error: %v", err)
	}

	want := &Build{BuildNum: 123}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Client.GetBuild(jszwedko, foo, 123) returned %+v, want %+v", build, want)
	}
}

func TestClient_ListBuildArtifacts(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/123/artifacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"Path": "/some/path"}]`)
	})

	artifacts, err := client.ListBuildArtifacts("jszwedko", "foo", 123)
	if err != nil {
		t.Errorf("Client.ListBuildArtifacts(jszwedko, foo, 123) returned error: %v", err)
	}

	want := []*Artifact{{Path: "/some/path"}}
	if !reflect.DeepEqual(artifacts, want) {
		t.Errorf("Client.ListBuildArtifacts(jszwedko, foo, 123) returned %+v, want %+v", artifacts, want)
	}
}

func TestClient_ListTestMetadata(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/123/tests", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"tests": [{"name": "some test"}]}`)
	})

	metadata, err := client.ListTestMetadata("jszwedko", "foo", 123)
	if err != nil {
		t.Errorf("Client.ListTestMetadata(jszwedko, foo, 123) returned error: %v", err)
	}

	want := []*TestMetadata{{Name: "some test"}}
	if !reflect.DeepEqual(metadata, want) {
		t.Errorf("Client.ListTestMetadata(jszwedko, foo, 123) returned %+v, want %+v", metadata, want)
	}
}

func TestClient_Build(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/tree/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"build_num": 123}`)
	})

	build, err := client.Build("jszwedko", "foo", "master")
	if err != nil {
		t.Errorf("Client.Build(jszwedko, foo, master) returned error: %v", err)
	}

	want := &Build{BuildNum: 123}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Client.Build(jszwedko, foo, master) returned %+v, want %+v", build, want)
	}
}

func TestClient_ParameterizedBuild(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/project/jszwedko/foo/tree/master", func(w http.ResponseWriter, r *http.Request) {
		testBody(t, r, `{"param":"foo"}`)
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"build_num": 123}`)
	})

	params := map[string]string{
		"param": "foo",
	}

	build, err := client.ParameterizedBuild("jszwedko", "foo", "master", params)
	if err != nil {
		t.Errorf("Client.Build(jszwedko, foo, master) returned error: %v", err)
	}

	want := &Build{BuildNum: 123}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Client.Build(jszwedko, foo, master) returned %+v, want %+v", build, want)
	}
}

func TestClient_RetryBuild(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/123/retry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"build_num": 124}`)
	})

	build, err := client.RetryBuild("jszwedko", "foo", 123)
	if err != nil {
		t.Errorf("Client.RetryBuild(jszwedko, foo, 123) returned error: %v", err)
	}

	want := &Build{BuildNum: 124}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Client.RetryBuild(jszwedko, foo, 123) returned %+v, want %+v", build, want)
	}
}

func TestClient_CancelBuild(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/123/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"build_num": 123}`)
	})

	build, err := client.CancelBuild("jszwedko", "foo", 123)
	if err != nil {
		t.Errorf("Client.CancelBuild(jszwedko, foo, 123) returned error: %v", err)
	}

	want := &Build{BuildNum: 123}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Client.CancelBuild(jszwedko, foo, 123) returned %+v, want %+v", build, want)
	}
}

func TestClient_ClearCache(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/build-cache", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"status": "cache cleared"}`)
	})

	status, err := client.ClearCache("jszwedko", "foo")
	if err != nil {
		t.Errorf("Client.ClearCache(jszwedko, foo) returned error: %v", err)
	}

	want := "cache cleared"
	if !reflect.DeepEqual(status, want) {
		t.Errorf("Client.ClearCache(jszwedko, foo) returned %+v, want %+v", status, want)
	}
}

func TestClient_AddEnvVar(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/envvar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"name":"bar","value":"baz"}`)
		fmt.Fprint(w, `{"name": "bar"}`)
	})

	status, err := client.AddEnvVar("jszwedko", "foo", "bar", "baz")
	if err != nil {
		t.Errorf("Client.AddEnvVar(jszwedko, foo, bar, baz) returned error: %v", err)
	}

	want := &EnvVar{Name: "bar"}
	if !reflect.DeepEqual(status, want) {
		t.Errorf("Client.AddEnvVar(jszwedko, foo, bar, baz) returned %+v, want %+v", status, want)
	}
}

func TestClient_ListEnvVars(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/envvar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testBody(t, r, "")
		fmt.Fprint(w, `[{"name": "bar", "value":"xxxbar"}]`)
	})

	status, err := client.ListEnvVars("jszwedko", "foo")
	if err != nil {
		t.Errorf("Client.ListEnvVars(jszwedko, foo) returned error: %v", err)
	}

	want := []EnvVar{
		{Name: "bar", Value: "xxxbar"},
	}

	if !reflect.DeepEqual(status, want) {
		t.Errorf("Client.ListEnvVars(jszwedko, foo) returned %+v, want %+v", status, want)
	}
}

func TestClient_DeleteEnvVar(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/envvar/bar", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteEnvVar("jszwedko", "foo", "bar")
	if err != nil {
		t.Errorf("Client.DeleteEnvVar(jszwedko, foo, bar) returned error: %v", err)
	}
}

func TestClient_AddSSHKey(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/project/jszwedko/foo/ssh-key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"hostname":"example.com","private_key":"some-key"}`)
		w.WriteHeader(http.StatusCreated)
	})

	err := client.AddSSHKey("jszwedko", "foo", "example.com", "some-key")
	if err != nil {
		t.Errorf("Client.AddSSHKey(jszwedko, foo, example.com, some-key) returned error: %v", err)
	}
}

func TestClient_GetActionOutput(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/some-s3-path", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `[{"Message":"hello"}, {"Message": "world"}]`)
	})

	action := &Action{HasOutput: true, OutputURL: server.URL + "/some-s3-path"}

	outputs, err := client.GetActionOutputs(action)
	if err != nil {
		t.Errorf("Client.GetActionOutput(%+v) returned error: %v", action, err)
	}

	want := []*Output{{Message: "hello"}, {Message: "world"}}
	if !reflect.DeepEqual(outputs, want) {
		t.Errorf("Client.GetActionOutput(%+v) returned %+v, want %+v", action, outputs, want)
	}
}

func TestClient_GetActionOutput_withDebug(t *testing.T) {
	setup()
	defer teardown()
	buf := bytes.NewBuffer(nil)
	client.Debug = true
	client.Logger = log.New(buf, "", 0)
	mux.HandleFunc("/some-s3-path", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `[{"Message":"hello"}, {"Message": "world"}]`)
	})

	action := &Action{HasOutput: true, OutputURL: server.URL + "/some-s3-path"}

	_, err := client.GetActionOutputs(action)
	if err != nil {
		t.Errorf("Client.GetActionOutput(%+v) returned error: %v", action, err)
	}

	output := buf.String()

	t.Logf("debug output:\n%s", output)
	if !strings.Contains(output, "request:") {
		t.Error(`expected "request:" to appear in debug output`)
	}
	if !strings.Contains(output, "HTTP/1.1") {
		t.Error(`expected http request to appear in debug output`)
	}

	if !strings.Contains(output, "response:") {
		t.Error(`expected "response:" to appear in debug output`)
	}
	if !strings.Contains(output, "HTTP/1.1 200 OK") {
		t.Error(`expected http request to appear in debug output`)
	}
}

func TestClient_GetActionOutput_noOutput(t *testing.T) {
	setup()
	defer teardown()

	action := &Action{HasOutput: false}

	outputs, err := client.GetActionOutputs(action)
	if err != nil {
		t.Errorf("Client.GetActionOutput(%+v) returned error: %v", action, err)
	}

	if outputs != nil {
		t.Errorf("Client.GetActionOutput(%+v) returned %+v: want %v", action, outputs, nil)
	}
}

func TestClient_ListCheckoutKeys(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/project/jszwedko/foo/checkout-key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `[{
			"public_key": "some public key",
			"type": "deploy-key",
			"fingerprint": "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2",
			"login": null,
			"preferred": true
		}]`)
	})

	checkoutKeys, err := client.ListCheckoutKeys("jszwedko", "foo")
	if err != nil {
		t.Errorf("Client.ListCheckoutKeys(jszwedko, foo) returned error: %v", err)
	}

	want := []*CheckoutKey{{
		PublicKey:   "some public key",
		Type:        "deploy-key",
		Fingerprint: "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2",
		Login:       nil,
		Preferred:   true,
	}}
	if !reflect.DeepEqual(checkoutKeys, want) {
		t.Errorf("Client.ListCheckoutKeys(jszwedko, foo) returned %+v, want %+v", checkoutKeys, want)
	}
}

func TestClient_CreateCheckoutKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/project/jszwedko/foo/checkout-key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"type":"github-user-key"}`)
		fmt.Fprintf(w, `{
			"public_key": "some public key",
			"type": "github-user-key",
			"fingerprint": "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2",
			"login": "jszwedko",
			"preferred": true
		}`)
	})

	checkoutKey, err := client.CreateCheckoutKey("jszwedko", "foo", "github-user-key")
	if err != nil {
		t.Errorf("Client.CreateCheckoutKey(jszwedko, foo, github-user-key) returned error: %v", err)
	}

	username := "jszwedko"
	want := &CheckoutKey{
		PublicKey:   "some public key",
		Type:        "github-user-key",
		Fingerprint: "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2",
		Login:       &username,
		Preferred:   true,
	}
	if !reflect.DeepEqual(checkoutKey, want) {
		t.Errorf("Client.Client.CreateCheckoutKey(jszwedko, foo, github-user-key) returned %+v, want %+v", checkoutKey, want)
	}
}

func TestClient_GetCheckoutKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/project/jszwedko/foo/checkout-key/37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
			"public_key": "some public key",
			"type": "deploy-key",
			"fingerprint": "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2",
			"login": null,
			"preferred": true
		}`)
	})

	checkoutKey, err := client.GetCheckoutKey("jszwedko", "foo", "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2")
	if err != nil {
		t.Errorf("Client.GetCheckoutKey(jszwedko, foo, 37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2) returned error: %v", err)
	}

	want := &CheckoutKey{
		PublicKey:   "some public key",
		Type:        "deploy-key",
		Fingerprint: "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2",
		Login:       nil,
		Preferred:   true,
	}
	if !reflect.DeepEqual(checkoutKey, want) {
		t.Errorf("Client.GetCheckoutKey(jszwedko, foo, 37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2) returned %+v, want %+v", checkoutKey, want)
	}
}

func TestClient_DeleteCheckoutKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/project/jszwedko/foo/checkout-key/37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprintf(w, `{"message": "ok"}`)
	})

	err := client.DeleteCheckoutKey("jszwedko", "foo", "37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2")
	if err != nil {
		t.Errorf("Client.DeleteCheckoutKey(jszwedko, foo, 37:27:f7:68:85:43:46:d2:e1:30:83:8f:f7:1b:ad:c2) returned error: %v", err)
	}
}

func TestClient_AddSSHUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/project/jszwedko/foo/123/ssh-users", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"ssh_users": [{"github_id": 1234, "login": "jszwedko"}]}`)
	})

	build, err := client.AddSSHUser("jszwedko", "foo", 123)
	if err != nil {
		t.Errorf("Client.AddSSHUser(jszwedko, foo, 123) returned error: %v", err)
	}

	want := &Build{SSHUsers: []*SSHUser{{GithubID: 1234, Login: "jszwedko"}}}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Client.AddSSHUser(jszwedko, foo, 123) returned %+v, want %+v", build, want)
	}
}

func TestClient_AddHerokuKey(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/heroku-key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"apikey":"53433a12-9c99-11e5-97f5-1458d009721"}`)
		fmt.Fprint(w, `""`)
	})

	err := client.AddHerokuKey("53433a12-9c99-11e5-97f5-1458d009721")
	if err != nil {
		t.Errorf("Client.AddHerokuKey(53433a12-9c99-11e5-97f5-1458d009721) returned error: %v", err)
	}
}
