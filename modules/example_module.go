package modules

import (
	"context"
	"flag"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

type ExampleModule struct {
	outputFile string
}

func (m *ExampleModule) Name() string {
	return "example"
}

func (m *ExampleModule) ParseFlags(args []string) error {
	fs := flag.NewFlagSet(m.Name(), flag.ContinueOnError)
	// Collect the output file name
	fs.StringVar(&m.outputFile, "output", "", "Output file name")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

func NewExampleModule() Module {
	return &ExampleModule{}
}

func (m *ExampleModule) Run(apiClient *datadog.APIClient, ctx context.Context) error {
	// Run the module
	return nil
}
