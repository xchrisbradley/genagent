package core

import (
	"reflect"
	"sync"
)

// Entity is a unique identifier for an agent
type Entity uint64

// Component is an interface that all components must implement
type Component interface{}

// System is an interface that all systems must implement
type System interface {
	Update(world *World, dt float64)
}

// World manages all entities, components, and systems
type World struct {
	mu sync.RWMutex

	entities     []Entity
	nextEntityID Entity

	// Components are stored in a map of component type to a map of entity to component
	components map[reflect.Type]map[Entity]Component

	// Systems that process entities and components
	systems []System

	// Component types that have been registered
	componentTypes []reflect.Type
}

// NewWorld creates a new ECS world
func NewWorld() *World {
	return &World{
		entities:       make([]Entity, 0),
		components:     make(map[reflect.Type]map[Entity]Component),
		systems:        make([]System, 0),
		componentTypes: make([]reflect.Type, 0),
	}
}

// RegisterComponent registers a new component type
func (w *World) RegisterComponent(componentType reflect.Type) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.components[componentType]; !exists {
		w.components[componentType] = make(map[Entity]Component)
		w.componentTypes = append(w.componentTypes, componentType)
	}
}

// RegisterSystem adds a system to the world
func (w *World) RegisterSystem(system System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.systems = append(w.systems, system)
}

// CreateEntity creates a new entity
func (w *World) CreateEntity() Entity {
	w.mu.Lock()
	defer w.mu.Unlock()

	entity := w.nextEntityID
	w.nextEntityID++
	w.entities = append(w.entities, entity)
	return entity
}

// AddComponent adds a component to an entity
func (w *World) AddComponent(entity Entity, component Component) {
	w.mu.Lock()
	defer w.mu.Unlock()

	componentType := reflect.TypeOf(component)
	if _, exists := w.components[componentType]; !exists {
		panic("Component type not registered: " + componentType.String())
	}

	w.components[componentType][entity] = component
}

// GetComponent returns a component for an entity
func (w *World) GetComponent(entity Entity, componentType reflect.Type) Component {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if components, exists := w.components[componentType]; exists {
		if component, ok := components[entity]; ok {
			return component
		}
	}
	return nil
}

// Update processes all systems
func (w *World) Update(dt float64) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, system := range w.systems {
		system.Update(w, dt)
	}
}

// Entities returns a slice of all entities
func (w *World) Entities() []Entity {
	w.mu.RLock()
	defer w.mu.RUnlock()

	entities := make([]Entity, len(w.entities))
	copy(entities, w.entities)
	return entities
}

// RemoveEntity removes an entity and all its components
func (w *World) RemoveEntity(entity Entity) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Remove all components
	for _, components := range w.components {
		delete(components, entity)
	}

	// Remove from entities slice
	for i, e := range w.entities {
		if e == entity {
			w.entities = append(w.entities[:i], w.entities[i+1:]...)
			break
		}
	}
}
