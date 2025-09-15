-- +goose Up
CREATE TABLE bid_decision (
    id SERIAL PRIMARY KEY,
    bid_id INT NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES employee(id) ON DELETE CASCADE,
    decision VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (bid_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS bid_decision;
