package modules

import (
	"context"
	"flag"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// To make a new module, rename this struct, any references to it, and the file name
type ExampleModule struct {
	outputFile string
}

// Name returns the name of the module.
// This is the name that the user will use to run the module.
func (m *ExampleModule) Name() string {
	return "example"
}

// ParseFlags parses the flags for the module.
// Put any command line options you want to support here.
// The options are saved in the module struct, above. Remmember to add them there too!
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

// Rename this method and add it to the bottom of modules/register.go to register your module.
func NewExampleModule() Module {
	return &ExampleModule{}
}

// All modules must implement the Run method. This is called from main.go as the entry point for the module.
func (m *ExampleModule) Run(apiClient *datadog.APIClient, ctx context.Context) error {
	// Run the module
	return nil
}
