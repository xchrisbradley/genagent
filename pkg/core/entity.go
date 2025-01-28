package core

import "fmt"

// Entity represents a basic unit in the world that can have components attached
type Entity interface {
	// ID returns the unique identifier of the entity
	ID() string

	// AddComponent adds a component to the entity
	AddComponent(component Component) error

	// RemoveComponent removes a component from the entity
	RemoveComponent(componentType string) error

	// GetComponent returns a component by its type
	GetComponent(componentType string) (Component, bool)

	// HasComponent checks if the entity has a specific component
	HasComponent(componentType string) bool

	// Components returns all components attached to the entity
	Components() []Component
}

// baseEntity provides a basic implementation of the Entity interface
type baseEntity struct {
	id         string
	components map[string]Component
}

// NewEntity creates a new entity instance
func NewEntity() Entity {
	return &baseEntity{
		id:         GenerateUUID(),
		components: make(map[string]Component),
	}
}

func (e *baseEntity) ID() string {
	return e.id
}

func (e *baseEntity) AddComponent(component Component) error {
	if component == nil {
		return fmt.Errorf("cannot add nil component")
	}
	e.components[component.Type()] = component
	return nil
}

func (e *baseEntity) RemoveComponent(componentType string) error {
	if _, exists := e.components[componentType]; !exists {
		return fmt.Errorf("component %s does not exist", componentType)
	}
	delete(e.components, componentType)
	return nil
}

func (e *baseEntity) GetComponent(componentType string) (Component, bool) {
	component, exists := e.components[componentType]
	return component, exists
}

func (e *baseEntity) HasComponent(componentType string) bool {
	_, exists := e.components[componentType]
	return exists
}

func (e *baseEntity) Components() []Component {
	components := make([]Component, 0, len(e.components))
	for _, component := range e.components {
		components = append(components, component)
	}
	return components
}
