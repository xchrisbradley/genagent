-- Recreate status enum and column
CREATE TYPE pipeline_status AS ENUM (
    'COMPLETED',
    'RUNNING',
    'FAILED',
    'CANCELLED',
    'TERMINATED',
    'TIMED_OUT',
    'CONTINUED_AS_NEW'
);

ALTER TABLE pipeline ADD COLUMN status pipeline_status NOT NULL DEFAULT 'RUNNING';
CREATE INDEX idx_pipeline_status ON pipeline(status);
