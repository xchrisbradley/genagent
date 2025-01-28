# GenAgent

GenAgent is a powerful agent-based system for task automation, providing a flexible and extensible framework for building autonomous systems with component-based architecture.

## Features

- Component-based architecture for building autonomous agents
- Plugin system for extending functionality
- Built-in logging and configuration management
- Easy-to-use CLI interface
- Real-time agent processing at 60 FPS
- Flexible host and port configuration

## Installation

```bash
# Clone the repository
git clone https://encore.app.git
cd genagent

# Install the project
go run main.go install
```

## Usage

### Basic Commands

```bash
# Show help and available commands
genagent --help

# Start the agent system (default: localhost:8080)
genagent start

# Start with custom host and port
genagent start --host 0.0.0.0 --port 9000
```

### Plugin Management

GenAgent supports plugin installation from both Git repositories and local paths:

```bash
# Install plugin from Git repository
genagent install --plugin-url https://github.com/user/plugin-repo

# Install plugin from local path
genagent install --plugin-path /path/to/local/plugin
```

## Configuration

GenAgent uses a configuration file located at `$HOME/.genagent.yaml`. You can specify a custom config file using the `--config` flag:

```bash
genagent --config /path/to/config.yaml
```

### Project Structure

GenAgent creates and manages the following directory structure:

```
.genagent/
├── logs/      # Log files for system operations
├── memory/    # Agent memory and state storage
├── storage/   # Persistent data storage
└── temp/      # Temporary files and caches
```

## Plugin Development

To create a new plugin for GenAgent:

1. Create a new directory for your plugin
2. Implement the plugin interface (see plugins/ai/plugin.go for example)
3. Add necessary components and systems
4. Test your plugin functionality
5. Install using the plugin installation commands

Example plugin structure:

```
plugin-name/
├── plugin.go      # Plugin interface implementation
├── components/    # Custom components
└── systems/       # Custom systems
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.