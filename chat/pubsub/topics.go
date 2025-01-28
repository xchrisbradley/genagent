package pubsub

import (
	"encore.app/chat/types"
	"encore.dev/pubsub"
)

// ChatEvents is the topic for all chat-related events
var ChatEvents = pubsub.NewTopic[*types.ChatEvent]("chat-events", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// ChatResponses is the topic for bot responses
var ChatResponses = pubsub.NewTopic[*types.ChatEvent]("chat-responses", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// PlatformEvents is the topic for platform-specific events (typing indicators, presence updates, etc)
var PlatformEvents = pubsub.NewTopic[*types.ChatEvent]("platform-events", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
