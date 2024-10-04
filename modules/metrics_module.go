package modules

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

type MetricsModule struct {
	outputFile string
	from       string
	to         string
	query      string
}

func (m *MetricsModule) Name() string {
	return "metrics"
}

func (m *MetricsModule) ParseFlags(args []string) error {
	// For the defaults for to and from, we use the current time and the time 1 hour hour ago.
	defaultTo := time.Now().Format(time.RFC3339)
	defaultFrom := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	fs := flag.NewFlagSet(m.Name(), flag.ContinueOnError)
	// Collect the output file name
	fs.StringVar(&m.outputFile, "output", "", "Output file name. If blank, output to stdout.")
	// Collect the from and to times
	fs.StringVar(&m.from, "from", defaultFrom, "From time: 31/12/2023, 31 Dec 2023...")
	fs.StringVar(&m.to, "to", defaultTo, "To time: 31/12/2023, 31 Dec 2023...")
	fs.StringVar(&m.query, "query", "avg:system.cpu.user{*}", "Query to run")

	// Capture the help flag
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", m.Name())
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// If help flag is provided, print usage and return an error to stop execution
	if fs.Parsed() && len(args) > 0 && args[0] == "--help" {
		fs.Usage()
		return flag.ErrHelp
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
	var output *os.File
	var osErr error
	if m.outputFile != "" {
		output, osErr = os.Create(m.outputFile)
		if osErr != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Writing output to %s\n", m.outputFile)
	} else {
		output = os.Stdout
	}
	defer output.Close()

	// TODO: Create headers in the CSV file based on the queries

	type MetricsData struct {
		Scope     string
		Timestamp float64
		Value     float64
	}
	var result []MetricsData

	// Run the query
	resp, err := RunQuery(ctx, metricsApi, fromTime, toTime, m.query)
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
			result = append(result, MetricsData{Scope: series.GetScope(), Timestamp: *point[0], Value: *pvalue})
		}
	}

	// Write the header to a file
	header := GetStructKeysAsCSV(MetricsData{})
	_, err = output.WriteString(header + "\n")
	if err != nil {
		return fmt.Errorf("error writing header to file: %v", err)
	}

	// Write the data rows
	for _, data := range result {
		_, err := output.WriteString(fmt.Sprintf("%s,%f,%f\n", data.Scope, data.Timestamp, data.Value))
		if err != nil {
			return err
		}
	}
	return nil
}
