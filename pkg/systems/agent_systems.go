package systems

import (
	"reflect"
	"time"

	"github.com/xchrisbradley/genagent/pkg/components"
	"github.com/xchrisbradley/genagent/pkg/core"
)

const (
	FLAG_ACTIVE uint32 = 1 << iota
	FLAG_LEARNING
	FLAG_PAUSED
	FLAG_ERROR
)

// ThinkingSystem processes cognitive operations for agents
type ThinkingSystem struct{}

func (s *ThinkingSystem) Update(world *core.World, dt float64) {
	thinkingType := reflect.TypeOf(&components.ThinkingComponent{})

	for _, entity := range world.Entities() {
		thinking := world.GetComponent(entity, thinkingType)
		if thinking == nil {
			continue
		}

		tc := thinking.(*components.ThinkingComponent)
		if tc.StateFlags&FLAG_PAUSED != 0 {
			continue
		}

		// Update thinking state
		tc.LastThought = time.Now()
		tc.Context["last_update"] = tc.LastThought
	}
}

// MemorySystem manages memory operations for agents
type MemorySystem struct {
	MemoryRetentionThreshold float64
}

func (s *MemorySystem) Update(world *core.World, dt float64) {
	memoryType := reflect.TypeOf(&components.MemoryComponent{})

	for _, entity := range world.Entities() {
		memory := world.GetComponent(entity, memoryType)
		if memory == nil {
			continue
		}

		mc := memory.(*components.MemoryComponent)
		mc.LastAccessed = time.Now()

		// Consolidate short-term to long-term memory
		s.consolidateMemories(mc)

		// Prune old memories based on importance
		s.pruneMemories(mc)
	}
}

func (s *MemorySystem) consolidateMemories(mc *components.MemoryComponent) {
	now := time.Now()
	var remainingShortTerm []components.Memory

	for _, memory := range mc.ShortTerm {
		// Move important memories to long-term
		if memory.Importance >= s.MemoryRetentionThreshold {
			mc.LongTerm = append(mc.LongTerm, memory)
		} else if now.Sub(memory.Timestamp) < 24*time.Hour {
			// Keep recent memories in short-term
			remainingShortTerm = append(remainingShortTerm, memory)
		}
	}

	mc.ShortTerm = remainingShortTerm
}

func (s *MemorySystem) pruneMemories(mc *components.MemoryComponent) {
	if len(mc.LongTerm) <= mc.Capacity {
		return
	}

	// Sort memories by importance and keep only up to capacity
	// Note: In a real implementation, you'd want to use a more sophisticated
	// algorithm that considers both importance and age
	mc.LongTerm = mc.LongTerm[len(mc.LongTerm)-mc.Capacity:]
}

// SensorSystem processes environmental inputs
type SensorSystem struct{}

func (s *SensorSystem) Update(world *core.World, dt float64) {
	sensorType := reflect.TypeOf(&components.SensorComponent{})

	for _, entity := range world.Entities() {
		sensor := world.GetComponent(entity, sensorType)
		if sensor == nil {
			continue
		}

		sc := sensor.(*components.SensorComponent)
		sc.LastUpdate = time.Now()

		// Process each active sensor
		for _, sensorName := range sc.ActiveSensors {
			if input, exists := sc.Inputs[sensorName]; exists {
				s.processSensorInput(sensorName, input)
			}
		}
	}
}

func (s *SensorSystem) processSensorInput(sensorName string, input interface{}) {
	// Implementation would depend on sensor type and input format
}

// ActuatorSystem manages agent actions
type ActuatorSystem struct{}

func (s *ActuatorSystem) Update(world *core.World, dt float64) {
	actuatorType := reflect.TypeOf(&components.ActuatorComponent{})

	for _, entity := range world.Entities() {
		actuator := world.GetComponent(entity, actuatorType)
		if actuator == nil {
			continue
		}

		ac := actuator.(*components.ActuatorComponent)

		// Process queued actions by priority
		s.processActions(ac)
	}
}

func (s *ActuatorSystem) processActions(ac *components.ActuatorComponent) {
	if len(ac.Actions) == 0 {
		return
	}

	// Sort actions by priority
	// Execute highest priority action
	action := ac.Actions[0]
	s.executeAction(action)

	// Remove executed action
	ac.Actions = ac.Actions[1:]
	ac.LastAction = time.Now()
}

func (s *ActuatorSystem) executeAction(action components.Action) {
	// Implementation would depend on action type and parameters
}

// CommunicationSystem handles agent communication
type CommunicationSystem struct{}

func (s *CommunicationSystem) Update(world *core.World, dt float64) {
	commType := reflect.TypeOf(&components.CommunicationComponent{})

	for _, entity := range world.Entities() {
		comm := world.GetComponent(entity, commType)
		if comm == nil {
			continue
		}

		cc := comm.(*components.CommunicationComponent)

		// Process message queue
		s.processMessages(cc)

		// Update channel states
		s.updateChannels(cc)
	}
}

func (s *CommunicationSystem) processMessages(cc *components.CommunicationComponent) {
	if len(cc.MessageQueue) == 0 {
		return
	}

	// Process messages by priority
	// Implementation would handle actual message transmission
	cc.LastCommunication = time.Now()
}

func (s *CommunicationSystem) updateChannels(cc *components.CommunicationComponent) {
	for _, channel := range cc.Channels {
		// Update channel status, handle reconnections, etc.
		_ = channel // Placeholder until implementation
	}
}
