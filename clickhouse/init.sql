CREATE DATABASE IF NOT EXISTS logs_db;
USE logs_db;

CREATE TABLE IF NOT EXISTS raw_logs (
    text String,
    level String,
    time DateTime,
    project_id UUID
) ENGINE = MergeTree()
ORDER BY (level, time);

CREATE TABLE IF NOT EXISTS filtered_logs_target (
    text String,
    level String,
    time DateTime,
    project_id UUID,
    original_level String
) ENGINE = MergeTree()
ORDER BY (level, time);

CREATE MATERIALIZED VIEW IF NOT EXISTS filtered_logs_mv TO filtered_logs_target AS
SELECT
    text,
    level,
    time,
    project_id,
    level AS original_level
FROM raw_logs
WHERE level IN ('warn', 'error');