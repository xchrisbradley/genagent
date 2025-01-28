-- Drop status related items since Temporal maintains workflow status
DROP INDEX idx_pipeline_status;
ALTER TABLE pipeline DROP COLUMN status;
DROP TYPE pipeline_status;
