-- +goose Up
CREATE TYPE tender_status AS ENUM ('Created', 'Published', 'Closed');
CREATE TYPE tender_service_type AS ENUM ('Construction', 'Delivery', 'Manufacture');

CREATE TABLE tender (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    service_type tender_service_type NOT NULL,
    status tender_status NOT NULL DEFAULT 'Created',
    organization_id INT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_tender_updated_at BEFORE UPDATE ON tender
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_tender_updated_at ON tender;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS tender;
DROP TYPE IF EXISTS tender_status;
DROP TYPE IF EXISTS tender_service_type;
