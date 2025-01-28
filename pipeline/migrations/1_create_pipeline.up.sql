-- Create enum type for pipeline status
CREATE TYPE pipeline_status AS ENUM (
    'COMPLETED',
    'RUNNING',
    'FAILED',
    'CANCELLED',
    'TERMINATED',
    'TIMED_OUT',
    'CONTINUED_AS_NEW'
);

-- Create pipeline table
CREATE TABLE pipeline (
    id SERIAL PRIMARY KEY, -- Changed from BIGSERIAL to SERIAL to match int in Go
    workflow_id TEXT NOT NULL,
    definition JSONB NOT NULL,
    status pipeline_status NOT NULL,
    submitted_date TIMESTAMP NOT NULL,
    completed_date TIMESTAMP
);

-- Add indexes for commonly queried fields
CREATE INDEX idx_pipeline_workflow_id ON pipeline(workflow_id);
CREATE INDEX idx_pipeline_status ON pipeline(status);
CREATE INDEX idx_pipeline_submitted_date ON pipeline(submitted_date DESC);
CREATE INDEX idx_pipeline_definition ON pipeline USING GIN (definition);
