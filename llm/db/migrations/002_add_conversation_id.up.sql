-- Add conversation_id column to llm_requests
ALTER TABLE llm_requests ADD COLUMN conversation_id TEXT NOT NULL DEFAULT '';
