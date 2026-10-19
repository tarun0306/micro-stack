CREATE TABLE IF NOT EXISTS products (
    id CHAR(27) PRIMARY KEY,
    name VARCHAR(57) NOT NULL,
    description VARCHAR(100) NOT NULL,
    price FLOAT
)