package pubsub

import (
	"encore.app/llm/types"
	"encore.dev/pubsub"
)

// GenerationRequests is a topic for LLM generation requests
var GenerationRequests = pubsub.NewTopic[*types.LLMRequestEvent]("llm-generation-requests", pubsub.TopicConfig{
	// Use at-least-once delivery since our handlers are idempotent
	// (we track request_id in database)
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// GenerationResponses is a topic for LLM generation responses
var GenerationResponses = pubsub.NewTopic[*types.LLMResponseEvent]("llm-generation-responses", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
