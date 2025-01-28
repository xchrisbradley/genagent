package pipeline

import (
	"encoding/json"
	"fmt"

	"encore.app/pipeline/nodes"
	"go.temporal.io/sdk/workflow"
)

// PipelineNode represents a node in the pipeline
type PipelineNode struct {
	ID     string          `json:"id"`
	Type   nodes.Type      `json:"type"`
	Config json.RawMessage `json:"config"`
	Next   []string        `json:"next"`
}

// PipelineDefinition represents a user-defined pipeline
type PipelineDefinition struct {
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Nodes       map[string]PipelineNode `json:"nodes"`
	EntryPoints []string                `json:"entryPoints"`
}

// ExecuteNode executes a single node in the pipeline
func executeNode(ctx workflow.Context, node PipelineNode) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing node", "id", node.ID, "type", node.Type)

	executor, ok := nodeRegistry.Get(node.Type)
	if !ok {
		return fmt.Errorf("unsupported node type: %s", node.Type)
	}

	result, err := executor.Execute(ctx, node.Config)
	if err != nil {
		return fmt.Errorf("failed to execute node %s: %w", node.ID, err)
	}

	if !result.Success {
		return fmt.Errorf("node %s failed: %s", node.ID, result.Error)
	}

	return nil
}

// ExecutePipline is the main workflow executor that handles user-defined pipelines
func ExecutePipeline(ctx workflow.Context, def PipelineDefinition) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting pipline execution", "name", def.Name, "version", def.Version)

	// Execute each entry point node
	for _, entryID := range def.EntryPoints {
		node, exists := def.Nodes[entryID]
		if !exists {
			return fmt.Errorf("entry point node not found: %s", entryID)
		}

		if err := executeNode(ctx, node); err != nil {
			return fmt.Errorf("failed to execute node %s: %w", entryID, err)
		}

		// Execute subsequent nodes
		for _, nextID := range node.Next {
			nextNode, exists := def.Nodes[nextID]
			if !exists {
				return fmt.Errorf("next node not found: %s", nextID)
			}

			if err := executeNode(ctx, nextNode); err != nil {
				return fmt.Errorf("failed to execute node %s: %w", nextID, err)
			}
		}
	}

	return nil
}
