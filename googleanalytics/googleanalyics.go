package googleanalytics

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	ga "google.golang.org/api/analyticsreporting/v4"
)

// Client holds the information for a Google Analytics reporting client.
type Client struct {
	config  *jwt.Config
	client  *http.Client
	service *ga.Service
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

	// Construct the analytics reporting service object.
	client.service, err = ga.New(client.client)
	if err != nil {
		return nil, fmt.Errorf("creating the analytics reporting service object failed: %v", err)
	}

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
					{Expression: "ga:users"},
				},
				Dimensions: []*ga.Dimension{
					//{Name: "ga:country"},
				},
			},
		},
	}

	// Call the BatchGet method and return the response.
	return c.service.Reports.BatchGet(req).Do()
}

// PrintResponse parses and prints the Analytics Reporting API V4 response.
func PrintResponse(resp *ga.GetReportsResponse) error {
	// Iterate over the reports.
	for _, report := range resp.Reports {
		header := report.ColumnHeader
		dimHdrs := header.Dimensions
		metricHdrs := header.MetricHeader.MetricHeaderEntries
		rows := report.Data.Rows

		if rows == nil {
			return fmt.Errorf("no data found for given view")
		}

		for _, row := range rows {
			dims := row.Dimensions
			metrics := row.Metrics

			for i := 0; i < len(dimHdrs) && i < len(dims); i++ {
				logrus.Infof("%s: %s", dimHdrs[i], dims[i])
			}

			for _, metric := range metrics {
				// We have only 1 date range in the example
				// So it'll always print "Date Range (0)"
				// log.Infof("Date Range (%d)", idx)
				for j := 0; j < len(metricHdrs) && j < len(metric.Values); j++ {
					logrus.Infof("%s: %s", metricHdrs[j].Name, metric.Values[j])
				}
			}
		}
	}

	return nil
}
