package googleanalytics

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	gav3 "google.golang.org/api/analytics/v3"
	ga "google.golang.org/api/analyticsreporting/v4"
)

const (
	gaPrefix = "ga:"
)

// Client holds the information for a Google Analytics reporting client.
type Client struct {
	config          *jwt.Config
	client          *http.Client
	service         *ga.Service
	servicev3       *gav3.Service
	realtimeService *gav3.DataRealtimeService
}

// New takes a keyfile for auththentication and
// returns a new Google Analytics Reporting Client struct.
// Your credentials should be obtained from the Google
// Developer Console (https://console.developers.google.com).
// Navigate to your project, then see the "Credentials" page
// under "APIs & Auth".
// To create a service account client, click "Create new Client ID",
// select "Service Account", and click "Create Client ID". A JSON
// key file will then be downloaded to your computer.
func New(keyfile string, debug bool) (*Client, error) {
	// Read the keyfile.
	data, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, fmt.Errorf("reading keyfile %q failed: %v", keyfile, err)
	}

	// Create the initial client.
	client := &Client{}

	// Create a JWT config from the keyfile.
	client.config, err = google.JWTConfigFromJSON(data, ga.AnalyticsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("creating JWT config from json keyfile %q failed: %v", keyfile, err)
	}

	// The following GET request will be authorized and authenticated
	// on the behalf of your service account.
	if debug {
		ctx := context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{Transport: &logTransport{http.DefaultTransport}},
		)
		client.client = client.config.Client(ctx)
	} else {
		client.client = client.config.Client(oauth2.NoContext)
	}

	// Construct the analytics reporting v4 service object.
	client.service, err = ga.New(client.client)
	if err != nil {
		return nil, fmt.Errorf("creating the analytics reporting service v4 object failed: %v", err)
	}

	// Construct the analytics reporting v3 service object.
	// TODO: remove v3 once v4 supports the realtime reporting API.
	client.servicev3, err = gav3.New(client.client)
	if err != nil {
		return nil, fmt.Errorf("creating the analytics reporting service v3 object failed: %v", err)
	}
	client.realtimeService = gav3.NewDataRealtimeService(client.servicev3)

	return client, nil
}

// GetReport queries the Analytics Reporting API V4 using the
// Analytics Reporting API V4 service object.
// It returns the Analytics Reporting API V4 response
func (c *Client) GetReport(viewID string) (*ga.GetReportsResponse, error) {
	req := &ga.GetReportsRequest{
		ReportRequests: []*ga.ReportRequest{
			{
				ViewId: viewID,
				DateRanges: []*ga.DateRange{
					// TODO(jessfraz): this should be pased into this function.
					{StartDate: "7daysAgo", EndDate: "today"},
				},
				Metrics: []*ga.Metric{
					{Expression: "ga:sessions"},
					{Expression: "ga:pageviews"},
					{Expression: "ga:uniquePageviews"},
					{Expression: "ga:users"},
				},
				Dimensions: []*ga.Dimension{
					{Name: "ga:pagePath"},
				},
				OrderBys: []*ga.OrderBy{
					{FieldName: "ga:sessions", SortOrder: "DESCENDING"},
					{FieldName: "ga:pageviews", SortOrder: "DESCENDING"},
				},
			},
		},
	}

	// Call the BatchGet method and return the response.
	return c.service.Reports.BatchGet(req).Do()
}

// GetRealtimeActiveUsers queries the Analytics Realtime Reporting API V3 using the
// Analytics Reporting API V3 service object.
// It returns the Analytics Realtime Reporting API V3 response
// for how many active users are currently on the site.
func (c *Client) GetRealtimeActiveUsers(viewID string) (string, error) {
	metric := "rt:activeUsers"

	// Call the realtime get method.
	resp, err := c.realtimeService.Get(gaPrefix+viewID, metric).Do()
	if err != nil {
		return "", err
	}

	return resp.TotalsForAllResults[metric], nil
}

// PrintResponse parses and prints the Analytics Reporting API V4 response
// in the form of a tabwriter table.
// It will only print X maxRows if passed. If 0 is passed for maxRows
// it will print all the rows.
func PrintResponse(resp *ga.GetReportsResponse, maxRows int) error {
	// Iterate over the reports.
	for _, report := range resp.Reports {
		if report.Data.Rows == nil {
			return fmt.Errorf("no data found for given view")
		}

		// Set the maxium rows to print. If it is 0, ie. the user did not pass one,
		// the set it to the length og the rows.
		if maxRows == 0 {
			maxRows = len(report.Data.Rows)
		}

		// Clean the dimensions headers.
		dimensionsHeaders := []string{}
		for a := 0; a < len(report.ColumnHeader.Dimensions); a++ {
			dimensionsHeaders = append(dimensionsHeaders, strings.TrimPrefix(report.ColumnHeader.Dimensions[a], gaPrefix))
		}

		// Clean the metric headers.
		metricHeaders := []string{}
		for i := 0; i < len(report.ColumnHeader.MetricHeader.MetricHeaderEntries); i++ {
			metricHeaders = append(metricHeaders, strings.TrimPrefix(report.ColumnHeader.MetricHeader.MetricHeaderEntries[i].Name, gaPrefix))
		}

		// Create the tabwriter.
		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

		// Print dimensions and metrics header.
		fmt.Fprintf(w, "%s\n", strings.ToUpper(strings.Join(append(dimensionsHeaders, metricHeaders...), "\t")))

		for l := 0; l < maxRows && l < len(report.Data.Rows); l++ {
			// Clean the metric values.
			values := []string{}
			for _, m := range report.Data.Rows[l].Metrics {
				for j := 0; j < len(m.Values); j++ {
					values = append(values, m.Values[j])
				}
			}

			// Print the dimensions and metrics.
			fmt.Fprintf(w, "%s\n", strings.Join(append(report.Data.Rows[l].Dimensions, values...), "\t"))
		}

		// Print the totals _only_ if we had dimensions.
		if len(report.ColumnHeader.Dimensions) > 0 {
			// Clean the dimensions headers for the totals row.
			headers := []string{}
			for h := 0; h < len(report.ColumnHeader.Dimensions); h++ {
				if h == 0 {
					headers = append(headers, "TOTAL")
					continue
				}
				headers = append(headers, "-")
			}

			// Clean the totals values.
			totals := []string{}
			for _, t := range report.Data.Totals {
				for k := 0; k < len(t.Values); k++ {
					totals = append(totals, t.Values[k])
				}
			}

			fmt.Fprintf(w, "%s\n", strings.Join(append(headers, totals...), "\t"))
		}

		w.Flush()
	}

	return nil
}

// getAccounts queries the Analytics Managemnt API V3 using the
// Analytics Management API V3 service object.
// It returns an array of analytics accounts.
func (c *Client) getAccounts() ([]*gav3.Account, error) {
	resp, err := gav3.NewManagementAccountsService(c.servicev3).List().Do()
	if err != nil {
		return nil, fmt.Errorf("listing accounts failed: %v", err)
	}

	return resp.Items, nil
}

// getProperties queries the Analytics Managemnt API V3 using the
// Analytics Management API V3 service object.
// It returns an array of analytics properties for an account ID.
func (c *Client) getProperties(accountID string) ([]*gav3.Webproperty, error) {
	resp, err := gav3.NewManagementWebpropertiesService(c.servicev3).List(accountID).Do()
	if err != nil {
		return nil, fmt.Errorf("listing properties failed: %v", err)
	}

	return resp.Items, nil
}

// getProfiles queries the Analytics Managemnt API V3 using the
// Analytics Management API V3 service object.
// It returns an array of analytics profiles for an account and property ID.
func (c *Client) getProfiles(accountID, propertyID string) ([]*gav3.Profile, error) {
	resp, err := gav3.NewManagementProfilesService(c.servicev3).List(accountID, propertyID).Do()
	if err != nil {
		return nil, fmt.Errorf("listing profiles failed: %v", err)
	}

	return resp.Items, nil
}

// GetProfileName returns the name of a Google Analytics profile.
func (c *Client) GetProfileName(profileID string) (name string, err error) {
	// Get the accounts.
	accounts, err := c.getAccounts()
	if err != nil {
		return "", err
	}

	// For each account get the properties.
	for _, account := range accounts {
		properties, err := c.getProperties(account.Id)
		if err != nil {
			return "", err
		}

		// Iterate over the properties
		for _, property := range properties {
			// Check early if the default profile is our profileID.
			// Then we won't have to do a call to getProfiles.
			if strconv.Itoa(int(property.DefaultProfileId)) == profileID {
				name = property.Name
				break
			}

			// Otherwise get the profiles for the property to find a match.
			profiles, err := c.getProfiles(account.Id, property.Id)
			if err != nil {
				return "", err
			}

			// Iterate over the profiles.
			for _, profile := range profiles {
				if profile.Id == profileID {
					name = profile.Name
					break
				}
			}
		}
	}

	return name, err
}
