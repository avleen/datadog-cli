package modules

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// To make a new module, rename this struct, any references to it, and the file name
type ContainersModule struct {
	outputFile        string
	containerGrouping string
	groupingKey       string
}

// Name returns the name of the module.
// This is the name that the user will use to run the module.
func (m *ContainersModule) Name() string {
	return "containers"
}

// ParseFlags parses the flags for the module.
// Put any command line options you want to support here.
// The options are saved in the module struct, above. Remmember to add them there too!
func (m *ContainersModule) ParseFlags(args []string) error {
	fs := flag.NewFlagSet(m.Name(), flag.ContinueOnError)
	// Collect the output file name
	fs.StringVar(&m.outputFile, "output", "output.json", "Output file name")
	fs.StringVar(&m.containerGrouping, "grouping", "ungrouped", "Specify the type of containers (ungrouped|grouped)")
	fs.StringVar(&m.groupingKey, "grouping-key", "image_name", "Specify the key to group containers by")
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	// Check the value of the containers flag
	if m.containerGrouping != "ungrouped" && m.containerGrouping != "grouped" {
		fmt.Fprintf(os.Stderr, "Invalid value for containers flag: %s\n", m.containerGrouping)
		os.Exit(1)
	}

	return nil
}

// Rename this method and add it to the bottom of modules/register.go to register your module.
func NewContainersModule() Module {
	return &ContainersModule{}
}

// All modules must implement the Run method. This is called from main.go as the entry point for the module.
func (m *ContainersModule) Run(apiClient *datadog.APIClient, ctx context.Context) error {
	// Run the module
	containersApi := datadogV2.NewContainersApi(apiClient)
	optionalParameters := datadogV2.NewListContainersOptionalParameters()
	optionalParameters.WithPageSize(1000)
	if m.containerGrouping == "grouped" {
		// Create a channel to unmarshal the results
		optionalParameters.WithGroupBy(m.groupingKey)
	}

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
	count := 0
	// Note the start time
	start := time.Now()

	// Call the API
	resp, _ := containersApi.ListContainersWithPagination(ctx, *optionalParameters)
	for paginationResult := range resp {
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
			fmt.Printf("Processed %d records at %.2f records/second\n", count, float64(count)/elapsed.Seconds())
		}
	}
	return nil
}

func (m *ContainersModule) processResults(results chan []byte) {
	file, err := os.Create(m.outputFile)
	defer file.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	for result := range results {
		// Write the JSON string to the file
		file.Write(result)
		file.Write([]byte("\n"))
		file.Sync()
	}
}
