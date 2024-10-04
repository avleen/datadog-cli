package main

import (
	"context"
	"datadog-cli/modules"
	"datadog-cli/pkg/config"
	"log"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Register modules
	// Example: RegisterModule(NewExampleModule())
	modules.RegisterAllModules()

	// Check if there are enough arguments
	if len(os.Args) < 2 {
		// Get the list of modules
		modules := modules.GetModules()
		log.Fatalf("No module specified. Available modules: %v", modules)
	}

	// Extract the module name
	moduleName := os.Args[1]

	// Find the module
	module, exists := modules.GetModule(moduleName)
	if !exists {
		log.Fatalf("Module not found: %s", moduleName)
	}

	// Parse the module-specific flags
	err = module.ParseFlags(os.Args[2:])
	if err != nil {
		// log.Fatalf("Error parsing flags for module %s: %v", moduleName, err)
		log.Fatalln("Exiting")
	}

	// Create a datadog client
	// Create a new Datadog client
	ctx := context.WithValue(context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: cfg.APIKey,
			},
			"appKeyAuth": {
				Key: cfg.APPKey,
			},
		})
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	// Run the module
	err = module.Run(apiClient, ctx)
	if err != nil {
		log.Fatalf("Error running module %s: %v", moduleName, err)
	}
}
