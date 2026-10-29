CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(27) PRIMARY KEY,
    name VARCHAR(57) NOT NULL,
    description VARCHAR(100) NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
    ) STORED
);

CREATE INDEX IF NOT EXISTS products_search_idx
    ON products USING GIN (search_vector);

CREATE INDEX IF NOT EXISTS products_name_trgm_idx
    ON products USING GIN (name gin_trgm_ops);