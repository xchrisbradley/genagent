-- Remove conversation_id column from llm_requests
ALTER TABLE llm_requests DROP COLUMN conversation_id;
