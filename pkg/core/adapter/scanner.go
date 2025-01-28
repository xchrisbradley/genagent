package adapter

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"encore.app/pkg/plugin"
)

// PluginScanner analyzes repository structure and detects potential plugin opportunities
type PluginScanner struct {
	basePath   string
	ignoreDirs []string
}

// NewPluginScanner creates a new scanner instance
func NewPluginScanner(basePath string) *PluginScanner {
	return &PluginScanner{
		basePath:   basePath,
		ignoreDirs: []string{".git", "node_modules", "vendor", ".genagent"},
	}
}

// ScanResult contains information about potential plugin opportunities
type ScanResult struct {
	Path         string
	Type         string
	Dependencies []string
	Metadata     plugin.PluginMetadata
}

// Scan walks through the repository and identifies potential plugin opportunities
func (ps *PluginScanner) Scan() ([]ScanResult, error) {
	var results []ScanResult

	err := filepath.Walk(ps.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if info.IsDir() {
			for _, ignore := range ps.ignoreDirs {
				if strings.Contains(path, ignore) {
					return filepath.SkipDir
				}
			}
		}

		// Detect potential plugin opportunities based on file patterns
		if !info.IsDir() {
			switch {
			case strings.HasSuffix(path, "go.mod"):
				result := ps.analyzeGoModule(path)
				if result != nil {
					results = append(results, *result)
				}
			case strings.HasSuffix(path, "package.json"):
				result := ps.analyzeNodePackage(path)
				if result != nil {
					results = append(results, *result)
				}
			}
		}

		return nil
	})

	return results, err
}

// analyzeGoModule examines a Go module for plugin potential
func (ps *PluginScanner) analyzeGoModule(path string) *ScanResult {
	// Read go.mod file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Extract module dependencies
	deps := []string{}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "require") {
			deps = append(deps, strings.TrimSpace(strings.TrimPrefix(line, "require")))
		}
	}

	return &ScanResult{
		Path:         path,
		Type:         "go",
		Dependencies: deps,
		Metadata: plugin.PluginMetadata{
			Description: "Go module with " + strconv.Itoa(len(deps)) + " dependencies",
			Tags:        []string{"go", "auto-detected"},
		},
	}
}

// analyzeNodePackage examines a Node.js package for plugin potential
func (ps *PluginScanner) analyzeNodePackage(path string) *ScanResult {
	// TODO: Implement Node.js package analysis
	return &ScanResult{
		Path: path,
		Type: "node",
		Metadata: plugin.PluginMetadata{
			Description: "Auto-detected Node.js package",
			Tags:        []string{"node", "auto-detected"},
		},
	}
}
