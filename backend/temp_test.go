package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	fmt.Println("Current LATEST_PROTOCOL_VERSION:", mcp.LATEST_PROTOCOL_VERSION)

	// Check if the new protocol version is supported
	fmt.Println("Valid protocol versions:")
	for i, v := range mcp.ValidProtocolVersions {
		fmt.Printf("  %d: %s\n", i+1, v)
	}

	// Test if "2025-11-25" is now supported
	expected := "2025-11-25"
	found := false
	for _, v := range mcp.ValidProtocolVersions {
		if v == expected {
			found = true
			break
		}
	}

	if found {
		fmt.Printf("\n✓ Success: %s is now supported!\n", expected)
	} else {
		fmt.Printf("\n✗ Error: %s is not supported\n", expected)
		os.Exit(1)
	}
}
