package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xchrisbradley/genagent/pkg/core/adapter"
)

var adaptCmd = &cobra.Command{
	Use:   "adapt [path]",
	Short: "Adapt an existing repository to use GenAgent",
	Long: `Adapt transforms an existing repository into a GenAgent-compatible project.
	It analyzes the repository structure, detects potential plugin opportunities,
	and sets up the necessary GenAgent infrastructure while preserving existing functionality.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get repository path (default to current directory)
		basePath := "."
		if len(args) > 0 {
			basePath = args[0]
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(basePath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %v", err)
		}

		// Create repository adapter
		repoAdapter, err := adapter.NewRepoAdapter(absPath)
		if err != nil {
			return fmt.Errorf("failed to create repository adapter: %v", err)
		}

		// Initialize GenAgent structure
		if err := repoAdapter.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize GenAgent structure: %v", err)
		}

		// Create and run scanner
		scanner := adapter.NewPluginScanner(absPath)
		results, err := scanner.Scan()
		if err != nil {
			return fmt.Errorf("failed to scan repository: %v", err)
		}

		// Register detected plugins
		for _, result := range results {
			repoAdapter.AddPlugin(result.Path)
			fmt.Printf("Detected potential plugin: %s (%s)\n", result.Path, result.Type)
		}

		// Start the adapter
		if err := repoAdapter.Start(); err != nil {
			return fmt.Errorf("failed to start adapter: %v", err)
		}

		fmt.Println("Successfully adapted repository to use GenAgent!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(adaptCmd)
}
