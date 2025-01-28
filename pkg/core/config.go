package core

import (
	"os"
	"path/filepath"
)

// ResourceConfig holds configuration for resource management
type ResourceConfig struct {
	BaseDir string
	Dirs    map[string]string
}

// NewResourceConfig creates a new resource configuration
func NewResourceConfig(projectRoot string) *ResourceConfig {
	baseDir := filepath.Join(projectRoot, ".genagent")

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		panic("Failed to create .genagent directory: " + err.Error())
	}

	// Define resource directories
	return &ResourceConfig{
		BaseDir: baseDir,
		Dirs: map[string]string{
			"logs":    filepath.Join(baseDir, "logs"),
			"memory":  filepath.Join(baseDir, "memory"),
			"storage": filepath.Join(baseDir, "storage"),
			"temp":    filepath.Join(baseDir, "temp"),
			"plugins": filepath.Join(baseDir, "plugins"),
		},
	}
}

// GetResourcePath returns the full path for a resource type
func (rc *ResourceConfig) GetResourcePath(resourceType string) string {
	if path, exists := rc.Dirs[resourceType]; exists {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(path, 0755); err != nil {
			panic("Failed to create resource directory: " + err.Error())
		}
		return path
	}
	return rc.BaseDir
}
