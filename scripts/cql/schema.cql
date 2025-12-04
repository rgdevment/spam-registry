CREATE KEYSPACE IF NOT EXISTS gsr 
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE gsr;

CREATE TABLE IF NOT EXISTS reports (
    id uuid,
    phone_number text,
    country_code text,
    reporter_hash text,
    category text,
    comment text,
    created_at timestamp,
    PRIMARY KEY ((phone_number), created_at)
) WITH CLUSTERING ORDER BY (created_at DESC)
  AND default_time_to_live = 47304000;

CREATE TABLE IF NOT EXISTS scores (
    phone_number text PRIMARY KEY,
    country_code text,
    score double,
    risk_level text,
    last_activity timestamp,
    velocity_hit_count int,
    total_reports int
) WITH default_time_to_live = 47304000;

CREATE TABLE IF NOT EXISTS active_threats (
    country_code text,
    risk_level text,
    phone_number text,
    score double,
    last_updated timestamp,
    PRIMARY KEY ((country_code), risk_level, phone_number)
) WITH default_time_to_live = 47304000;