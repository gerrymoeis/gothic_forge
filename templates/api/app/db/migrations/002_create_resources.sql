-- +goose Up
-- Create example resources table (customize for your API)
CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    data JSONB,
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_resources_name ON resources(name);
CREATE INDEX idx_resources_created_by ON resources(created_by);
CREATE INDEX idx_resources_data ON resources USING GIN(data);

-- +goose Down
DROP TABLE IF EXISTS resources;
