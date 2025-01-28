package components

import "time"

// ThinkingComponent represents an agent's cognitive capabilities
type ThinkingComponent struct {
	Confidence  float64
	StateFlags  uint32
	LastThought time.Time
	Context     map[string]interface{}
}

// MemoryComponent represents an agent's memory capabilities
type MemoryComponent struct {
	ShortTerm    []Memory
	LongTerm     []Memory
	Capacity     int
	LastAccessed time.Time
}

// Memory represents a single memory entry
type Memory struct {
	Content    interface{}
	Timestamp  time.Time
	Importance float64
	Tags       []string
}

// SensorComponent represents an agent's ability to perceive its environment
type SensorComponent struct {
	Inputs        map[string]interface{}
	LastUpdate    time.Time
	ActiveSensors []string
}

// ActuatorComponent represents an agent's ability to take actions
type ActuatorComponent struct {
	Actions      []Action
	LastAction   time.Time
	Capabilities []string
}

// Action represents a single action that can be taken
type Action struct {
	Name       string
	Parameters map[string]interface{}
	Priority   float64
	Status     string
}

// CommunicationComponent represents an agent's ability to communicate
type CommunicationComponent struct {
	Channels          []Channel
	MessageQueue      []Message
	LastCommunication time.Time
}

// Channel represents a communication channel
type Channel struct {
	ID     string
	Type   string
	Status string
	Config map[string]interface{}
}

// Message represents a communication message
type Message struct {
	ID        string
	Content   interface{}
	Timestamp time.Time
	Channel   string
	Priority  float64
}

// ResourceComponent represents an agent's resource management
type ResourceComponent struct {
	CPU       float64
	Memory    float64
	Storage   float64
	Network   float64
	LastCheck time.Time
}
