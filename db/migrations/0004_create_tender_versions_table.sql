-- +goose Up
CREATE TYPE tender_status AS ENUM ('Created', 'Published', 'Closed');
CREATE TYPE tender_service_type AS ENUM ('Construction', 'Delivery', 'Manufacture');

CREATE TABLE tender_versions (
    id SERIAL PRIMARY KEY,
    tender_id INT NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    service_type tender_service_type NOT NULL,
    status tender_status NOT NULL,
    organization_id INT NOT NULL,
    version INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE tender_versions;
