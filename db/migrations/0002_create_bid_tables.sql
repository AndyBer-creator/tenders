-- +goose Up
CREATE TABLE bid (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL,
    tender_id INT NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
    organization_id INT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    creator_username VARCHAR(50) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bid_versions (
    id SERIAL PRIMARY KEY,
    bid_id INT NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL,
    tender_id INT NOT NULL,
    organization_id INT NOT NULL,
    creator_username VARCHAR(50) NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE IF EXISTS bid_versions;
DROP TABLE IF EXISTS bid;
