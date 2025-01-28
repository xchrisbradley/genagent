package cmd

import (
	"os"
	"path/filepath"

	"encore.app/pkg/core"
	"github.com/spf13/cobra"
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
     - Useful for network access`,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize color scheme
			colors := core.DefaultColorScheme()

			// Initialize logger
			logger, err := core.NewLogger(filepath.Join(".genagent", "logs"), true)
			if err != nil {
				colors.Error("Error initializing logger: %v", err)
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

			logger.Info("Starting agent system on %s:%d", host, port)
			logger.Debug("Initializing world...")
			world := core.NewWorld()

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
