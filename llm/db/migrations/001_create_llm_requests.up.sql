-- Create llm_requests table
CREATE TABLE llm_requests (
    request_id TEXT PRIMARY KEY,
    bot_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    messages JSONB NOT NULL,
    parameters JSONB NOT NULL,
    response TEXT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE
);
