-- Create enum type for policy status
CREATE TYPE policy_status AS ENUM (
    'COMPLETED',
    'RUNNING',
    'FAILED',
    'CANCELLED',
    'TERMINATED',
    'TIMED_OUT',
    'CONTINUED_AS_NEW'
);

-- Create policy table
CREATE TABLE policy (
    id SERIAL PRIMARY KEY, -- Changed from BIGSERIAL to SERIAL to match int in Go
    workflow_id TEXT NOT NULL,
    definition JSONB NOT NULL,
    status policy_status NOT NULL,
    submitted_date TIMESTAMP NOT NULL,
    completed_date TIMESTAMP
);

-- Add indexes for commonly queried fields
CREATE INDEX idx_policy_workflow_id ON policy(workflow_id);
CREATE INDEX idx_policy_status ON policy(status);
CREATE INDEX idx_policy_submitted_date ON policy(submitted_date DESC);
CREATE INDEX idx_policy_definition ON policy USING GIN (definition);
