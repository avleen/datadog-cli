package modules

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// To make a new module, rename this struct, any references to it, and the file name
type HostsModule struct {
	outputFile string
	limit      int64
}

// Name returns the name of the module.
// This is the name that the user will use to run the module.
func (m *HostsModule) Name() string {
	return "hosts"
}

// ParseFlags parses the flags for the module.
// Put any command line options you want to support here.
// The options are saved in the module struct, above. Remmember to add them there too!
func (m *HostsModule) ParseFlags(args []string) error {
	fs := flag.NewFlagSet(m.Name(), flag.ContinueOnError)
	// Collect the output file name
	fs.StringVar(&m.outputFile, "output", "", "Output file name. If blank, output to stdout.")
	fs.Int64Var(&m.limit, "limit", 1000, "Limit the number of results returned. The API can currently return a maximum of 1000 results.")
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	return nil
}

// Rename this method and add it to the bottom of modules/register.go to register your module.
func NewHostsModule() Module {
	return &HostsModule{}
}

// All modules must implement the Run method. This is called from main.go as the entry point for the module.
func (m *HostsModule) Run(apiClient *datadog.APIClient, ctx context.Context) error {
	// Run the module
	hostsApi := datadogV1.NewHostsApi(apiClient)
	optionalParameters := datadogV1.NewListHostsOptionalParameters()
	optionalParameters.WithCount(m.limit)

	// Create a wait group to wait for the goroutines to finish
	var wg sync.WaitGroup
	// Create a channel to receive the results
	resultsChan := make(chan []byte)
	defer close(resultsChan)
	// Start a goroutine to process the results
	wg.Add(1)
	go func() {
		defer wg.Done()
		go m.processResults(resultsChan)
	}()

	// Count the number of records processed
	// count := 0
	// Note the start time
	// start := time.Now()

	// Call the API
	resp, httpResp, _ := hostsApi.ListHosts(ctx, *optionalParameters)
	// The following code was copied from the containers module. Unfortunately the DataDog API for ListHosts has no pagination.
	// Hopefully one day it will and we can use this code.
	/* for paginationResult := range resp {
		if paginationResult.Error != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ContainersApi.ListContainersWithPagination`: %v\n", paginationResult.Error)
		}
		// Send the entire result to the channel
		// Convert the item to a JSON string
		// jsonItem, _ := json.MarshalIndent(item, "", "  ")
		jsonItem, _ := json.Marshal(paginationResult.Item)
		// Send the JSON string to the results channel
		resultsChan <- jsonItem
		count++
		// Every 10,000 records, print a message with the count and the records/second since start
		if count%10000 == 0 {
			elapsed := time.Since(start)
			fmt.Fprintf(os.Stderr, "Processed %d records at %.2f records/second\n", count, float64(count)/elapsed.Seconds())
		}
	} */
	if httpResp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Error when calling `HostsApi.ListHosts`: %v\n", httpResp.Status)
	}
	jsonItem, _ := json.MarshalIndent(resp, "", "  ")
	resultsChan <- jsonItem
	// wait for the goroutine to finish
	wg.Wait()
	return nil
}

func (m *HostsModule) processResults(results chan []byte) {
	// TODO: For some reason writing the results to stdout is not working. It's not clear why.
	// I think there's a race condition somewhere.
	// Writing to a file is fine.
	var output *os.File
	var err error
	if m.outputFile != "" {
		output, err = os.Create(m.outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Writing output to %s\n", m.outputFile)
	} else {
		output = os.Stdout
	}
	defer output.Close()

	for result := range results {
		// Write the JSON string to the file
		output.Write(result)
		output.Write([]byte("\n"))
		output.Sync()
	}
}
