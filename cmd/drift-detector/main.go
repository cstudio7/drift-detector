package main

import (
	"fmt"
	"os"

	"github.com/cstudio7/drift-detector/internal/commands"
	"github.com/cstudio7/drift-detector/internal/third_party/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to load .env file: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	// Create a logger
	logger := logger.NewStdLogger()

	// Initialize command handler
	cmd := commands.NewDriftCommand(logger)

	// Check for command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [up|down|detect] [instance-id (for down)] [tfstate-file (for detect)]")
		fmt.Println("Example: go run main.go up")
		fmt.Println("Example: go run main.go down i-1234567890abcdef0")
		fmt.Println("Example: go run main.go detect terraform.tfstate")
		os.Exit(1)
	}

	// Run the command
	if err := cmd.Run(os.Args[1:]); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
