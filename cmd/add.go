package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [plugin-name]",
	Short: "Add a new plugin to GenAgent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pluginName := args[0]
		pluginDir := filepath.Join(".genagent", "plugins", pluginName)

		// Create plugin directory
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			fmt.Printf("Error creating plugin directory: %v\n", err)
			os.Exit(1)
		}

		// Create plugin.go file
		pluginFile := filepath.Join(pluginDir, "plugin.go")
		if _, err := os.Stat(pluginFile); err == nil {
			fmt.Printf("Plugin %s already exists\n", pluginName)
			os.Exit(1)
		}

		// Write plugin template
		template := fmt.Sprintf(`package %s

import (
	"fmt"

	"github.com/xchrisbradley/genagent/pkg/core"
)

type %sPlugin struct{}

func (p *%sPlugin) ID() string {
	return "%s"
}

func (p *%sPlugin) Name() string {
	return "%s Plugin"
}

func (p *%sPlugin) Version() string {
	return "1.0.0"
}

func (p *%sPlugin) Initialize(world *core.World, entity core.Entity) error {
	fmt.Printf("Initializing %s plugin...\n", p.Name())
	return nil
}

func (p *%sPlugin) Components() []core.Component {
	return nil
}

func (p *%sPlugin) Systems() []core.System {
	return nil
}

func (p *%sPlugin) Metadata() core.PluginMetadata {
	return core.PluginMetadata{
		Description: "A plugin for GenAgent",
		Author:      "GenAgent User",
		Website:     "https://github.com/xchrisbradley/genagent",
		Tags:        []string{"plugin"},
	}
}

// New creates a new instance of the plugin
func New() core.Plugin {
	return &%sPlugin{}
}
`,
			pluginName, pluginName, pluginName, pluginName, pluginName,
			pluginName, pluginName, pluginName, pluginName, pluginName,
			pluginName, pluginName, pluginName)

		if err := os.WriteFile(pluginFile, []byte(template), 0644); err != nil {
			fmt.Printf("Error writing plugin file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created plugin: %s\n", pluginName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
