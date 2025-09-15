-- +goose Up
CREATE TABLE bid_review (
    id SERIAL PRIMARY KEY,
    bid_id INT NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    description VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS bid_review;
