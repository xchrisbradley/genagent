package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xchrisbradley/genagent/pkg/core"
)

var (
	port     int
	host     string
	startCmd = &cobra.Command{
		Use:   "start [flags]",
		Short: "Start the agent system",
		Long: `Start command initializes and runs the agent system with the specified configuration.

Usage:
  1. Start with default settings:
     genagent start
     - Runs on localhost:8080
     - Uses default configuration

  2. Custom configuration:
     genagent start --host 0.0.0.0 --port 9000
     - Specify custom host and port
     - Useful for network access

The command sets up the world, registers components and systems, and begins
processing agent tasks. Use environment variables or config file to customize
further settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger
			logger, err := core.NewLogger(filepath.Join(".genagent", "logs"), true)
			if err != nil {
				fmt.Printf("Error initializing logger: %v\n", err)
				return
			}

			// Get project root
			wd, err := os.Getwd()
			if err != nil {
				logger.Error("Error getting working directory: %v", err)
				return
			}

			// Initialize resource configuration
			config := core.NewResourceConfig(wd)

			// Create resource directories if they don't exist
			for name, dir := range config.Dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					logger.Error("Error creating %s directory: %v", name, err)
					return
				}
			}

			// Download and update dependencies
			logger.Info("Updating dependencies...")
			goModTidy := exec.Command("go", "mod", "tidy")
			output, err := goModTidy.CombinedOutput()
			if err != nil {
				logger.Error("Error updating dependencies: %v\nOutput: %s", err, output)
				return
			}
			logger.Debug("Dependencies updated successfully")

			logger.Info("Starting agent system on %s:%d", host, port)
			logger.Debug("Initializing world...")
			world := core.NewWorld()

			// Load and initialize plugins
			logger.Info("Loading plugins...")
			pluginManager := core.NewPluginManager(world, config.Dirs["plugins"])
			if err := pluginManager.LoadPlugins(); err != nil {
				logger.Error("Error loading plugins: %v", err)
				return
			}

			logger.Info("Agent system is ready")

			// Main update loop
			for {
				world.Update(1.0 / 60.0) // 60 FPS
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(startCmd)

	// Local flags
	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the agent system on")
	startCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to run the agent system on")
}
