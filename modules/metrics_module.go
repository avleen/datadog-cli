package modules

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

type MetricsModule struct {
	outputFile string
	from       string
	to         string
}

func (m *MetricsModule) Name() string {
	return "metrics"
}

func (m *MetricsModule) ParseFlags(args []string) error {
	fs := flag.NewFlagSet(m.Name(), flag.ContinueOnError)
	// Collect the output file name
	fs.StringVar(&m.outputFile, "output", "", "Output file name")
	// Collect the from and to times
	fs.StringVar(&m.from, "from", "", "From time: 1 hour ago, 2 weeks ago..")
	fs.StringVar(&m.to, "to", "", "To time: now, 5 minutes ago..")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

func NewMetricsModule() Module {
	return &MetricsModule{}
}

func RunQuery(ctx context.Context, metricsApi *datadogV1.MetricsApi, from int64, to int64, query string) (datadogV1.MetricsQueryResponse, error) {
	// Fetch the metrics data
	resp, r, err := metricsApi.QueryMetrics(ctx, from, to, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.QueryMetrics`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	return resp, nil
}

func (m *MetricsModule) Run(apiClient *datadog.APIClient, ctx context.Context) error {
	// Run the module
	metricsApi := datadogV1.NewMetricsApi(apiClient)

	// Get the start and end times for the data
	fromTime, toTime, err := getToFrom(m.from, m.to)
	if err != nil {
		return err
	}

	// Create the output file
	filename := fmt.Sprintf(m.outputFile)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO: Create headers in the CSV file based on the queries

	type MetricsData struct {
		Scope     string
		Timestamp float64
		Value     float64
	}
	sums := make(map[string]MetricsData)

	// Run the query
	bytes_query := "sum:gcp.bigquery.storage.uploaded_bytes{*} by {project_id}.as_count()"
	resp, err := RunQuery(ctx, metricsApi, fromTime, toTime, bytes_query)
	if err != nil {
		return err
	}

	// For each series, collect the data points. Store the data points in a list.
	// Sum them up and print the result.
	for _, series := range resp.GetSeries() {
		for _, point := range series.GetPointlist() {
			if len(point) == 0 || point[1] == nil {
				continue
			}
			pvalue := point[1]
			if metric, exists := sums[series.GetScope()]; exists {
				metric.Value += *pvalue
			} else {
				sums[series.GetScope()] = MetricsData{Scope: series.GetScope(), Timestamp: *point[0], Value: *pvalue}
			}
		}
	}

	// Sort the list of sums and print the result
	var keys []string
	for k := range sums {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write the data to a file
	for _, scope := range keys {
		_, err := f.WriteString(fmt.Sprintf("%s,%f,%f\n", scope, sums[scope].Timestamp, sums[scope].Value))
		if err != nil {
			return err
		}
	}
	fmt.Printf("Data written to %s\n", filename)
	return nil
}
