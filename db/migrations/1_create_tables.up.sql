-- Create pipeline table
CREATE TABLE pipeline (
    id SERIAL PRIMARY KEY, -- Changed from BIGSERIAL to SERIAL to match int in Go
    workflow_id TEXT NOT NULL,
    definition JSONB NOT NULL,
    submitted_date TIMESTAMP NOT NULL,
    completed_date TIMESTAMP
);

-- Add indexes for commonly queried fields
CREATE INDEX idx_pipeline_workflow_id ON pipeline(workflow_id);
CREATE INDEX idx_pipeline_submitted_date ON pipeline(submitted_date DESC);
CREATE INDEX idx_pipeline_definition ON pipeline USING GIN (definition);

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
    completed_at TIMESTAMP WITH TIME ZONE,
	conversation_id TEXT NOT NULL DEFAULT '',
);

-- Create indexes
CREATE INDEX idx_llm_requests_bot_id ON llm_requests(bot_id);
CREATE INDEX idx_llm_requests_channel_id ON llm_requests(channel_id);
CREATE INDEX idx_llm_requests_created_at ON llm_requests(created_at);
CREATE INDEX idx_llm_requests_completed_at ON llm_requests(completed_at);

-- Create agent table
CREATE TABLE agent (
	id TEXT PRIMARY KEY,
	bio TEXT,
	profile_url TEXT
);

-- Create agent_embeddings table
CREATE TABLE agent_embeddings (
	id TEXT PRIMARY KEY REFERENCES agent(id),
	embedding REAL[] NOT NULL
);

-- Create agent_embeddings_index table
CREATE TABLE agent_embeddings_index (
	id SERIAL PRIMARY KEY,
	agent_id TEXT REFERENCES agent(id),
	embedding REAL[] NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_agent_embeddings_id ON agent_embeddings(id);
CREATE INDEX idx_agent_embeddings_index_agent_id ON agent_embeddings_index(agent_id);

-- Crate protocol_records table
CREATE TABLE protocol_records (
    id SERIAL PRIMARY KEY,
    url VARCHAR(2048) NOT NULL,
    protocol VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    response_time REAL NOT NULL,
    content TEXT,
    filename VARCHAR(255),
    downloaded_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create protocol_stats table
CREATE TABLE protocol_stats (
    protocol VARCHAR(10) PRIMARY KEY,
    total_requests INTEGER NOT NULL DEFAULT 0,
    successful_requests INTEGER NOT NULL DEFAULT 0,
    failed_requests INTEGER NOT NULL DEFAULT 0,
    average_response_time REAL NOT NULL DEFAULT 0,
    last_used TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_protocol_records_protocol ON protocol_records(protocol);
CREATE INDEX idx_protocol_records_downloaded_at ON protocol_records(downloaded_at);

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

CREATE TABLE "seed" (
	"id" serial PRIMARY KEY NOT NULL,
	"url" text NOT NULL,
	"domain" text NOT NULL,
	"priority" integer DEFAULT 1 NOT NULL,
	"next_crawl_after" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now(),
	"last_crawled_at" timestamp with time zone,
	CONSTRAINT "seed_url_unique" UNIQUE("url")
);

-- Create indexes
CREATE INDEX idx_seed_domain ON seed(domain);
CREATE INDEX "idx_seed_queue" ON "seed" USING btree ("domain","next_crawl_after");

CREATE TABLE "urls" (
	"url" text PRIMARY KEY NOT NULL,
	"first_seen_at" timestamp with time zone DEFAULT now(),
	"times_seen" integer DEFAULT 1
);

-- Create indexes
CREATE INDEX idx_urls_first_seen_at ON urls(first_seen_at);

CREATE TABLE "processing_records" (
	"id" serial PRIMARY KEY NOT NULL,
	"url" varchar(2048) NOT NULL,
	"processor_type" varchar(50) NOT NULL,
	"result" jsonb NOT NULL,
	"processed_at" timestamp NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);

-- Create indexes
CREATE INDEX idx_processing_records_url ON processing_records(url);

CREATE TABLE "dns_records" (
	"id" serial PRIMARY KEY NOT NULL,
	"hostname" varchar(256) NOT NULL,
	"ip_address" varchar(45) NOT NULL,
	"resolved_at" timestamp NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);

-- Create indexes
CREATE INDEX idx_dns_records_hostname ON dns_records(hostname);

CREATE TABLE IF NOT EXISTS "bronze" (
    "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "data" text NOT NULL,
    "source" text NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL
);

-- Create indexes
CREATE INDEX idx_bronze_created_at ON bronze(created_at);

CREATE TABLE IF NOT EXISTS "silver" (
    "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "raw_data_id" integer NOT NULL,
    "cleaned_data" text NOT NULL,
    "cleaning_method" text NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL
);

-- Create indexes
CREATE INDEX idx_silver_created_at ON silver(created_at);

CREATE TABLE IF NOT EXISTS "gold" (
    "id" INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "cleaned_data_id" integer NOT NULL,
    "enriched_data" text NOT NULL,
    "enrichment_method" text NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL
);
-- Create indexes
CREATE INDEX idx_gold_created_at ON gold(created_at);
