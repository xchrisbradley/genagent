package core

import "fmt"

// Integration represents a pluggable component that can be added to an agent
type Integration interface {
	// ID returns a unique identifier for this integration
	ID() string

	// Name returns a human-readable name for this integration
	Name() string

	// Version returns the semantic version of this integration
	Version() string

	// Initialize is called when the integration is first added to an agent
	Initialize(world *World, entity Entity) error

	// Components returns any components this integration needs to register
	Components() []Component

	// Systems returns any systems this integration needs to register
	Systems() []System
}

// IntegrationRegistry manages available integrations
type IntegrationRegistry struct {
	integrations map[string]Integration
}

// NewIntegrationRegistry creates a new integration registry
func NewIntegrationRegistry() *IntegrationRegistry {
	return &IntegrationRegistry{
		integrations: make(map[string]Integration),
	}
}

// Register adds an integration to the registry
func (r *IntegrationRegistry) Register(integration Integration) error {
	id := integration.ID()
	if _, exists := r.integrations[id]; exists {
		return fmt.Errorf("integration with ID %s already registered", id)
	}
	r.integrations[id] = integration
	return nil
}

// Get returns an integration by ID
func (r *IntegrationRegistry) Get(id string) (Integration, bool) {
	integration, exists := r.integrations[id]
	return integration, exists
}

// List returns all registered integrations
func (r *IntegrationRegistry) List() []Integration {
	var list []Integration
	for _, integration := range r.integrations {
		list = append(list, integration)
	}
	return list
}
