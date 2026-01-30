package main

import (
	"fmt"
	"os"

	"github.com/kymo-mcp/mcpcan/internal/market/app"
	"github.com/kymo-mcp/mcpcan/pkg/llm/models"
)

func main() {
	// Create application instance
	appInstance, err := app.New()
	if err != nil {
		fmt.Printf("Failed to create application instance: %v\n", err)
		os.Exit(1)
	}

	// Initialize application
	if err := appInstance.Initialize(); err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Log supported models info
	fmt.Printf("Supported LLM Providers: %d, Models: %d\n", len(models.AllProviders), len(models.GetAllModels()))

	// Run application
	if err := appInstance.Run(); err != nil {
		fmt.Printf("Failed to run application: %v\n", err)
		os.Exit(1)
	}
}
