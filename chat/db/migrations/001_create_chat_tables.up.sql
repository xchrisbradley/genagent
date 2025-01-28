-- Create bots table
CREATE TABLE bots (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    persona TEXT NOT NULL,
    avatar TEXT,
    provider VARCHAR(50) NOT NULL,
    parameters JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create conversations table
CREATE TABLE conversations (
    id VARCHAR(255) PRIMARY KEY,
    channel_id VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    bot_ids TEXT[] NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create messages table
CREATE TABLE messages (
    id VARCHAR(255) PRIMARY KEY,
    conversation_id VARCHAR(255) NOT NULL REFERENCES conversations(id),
    user_id VARCHAR(255) NOT NULL,
    bot_id VARCHAR(255) REFERENCES bots(id),
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_message_type CHECK (type IN ('text', 'image'))
);

-- Create indexes
CREATE INDEX idx_conversations_channel ON conversations(channel_id, platform);
CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
