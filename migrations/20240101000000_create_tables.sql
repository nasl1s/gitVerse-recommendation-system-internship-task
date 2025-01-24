-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    category VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recommendations (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    product_ids INT[] NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS statistics (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id),
    views INT NOT NULL DEFAULT 0,
    purchases INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    product_id INT NOT NULL REFERENCES products(id),
    purchased_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS likes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    product_id INT NOT NULL REFERENCES products(id),
    liked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_likes_per_user_product UNIQUE (user_id, product_id)
);

CREATE TABLE IF NOT EXISTS dislikes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    product_id INT NOT NULL REFERENCES products(id),
    disliked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_dislikes_per_user_product UNIQUE (user_id, product_id)
);

CREATE TABLE IF NOT EXISTS user_category_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    category VARCHAR(255) NOT NULL,
    score NUMERIC(10,2) NOT NULL DEFAULT 0,
    UNIQUE (user_id, category)
);

CREATE TABLE IF NOT EXISTS product_analytics (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL UNIQUE,
    likes INT NOT NULL DEFAULT 0,
    dislikes INT NOT NULL DEFAULT 0,
    purchases INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_analytics (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    total_likes INT NOT NULL DEFAULT 0,
    total_dislikes INT NOT NULL DEFAULT 0,
    total_purchases INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users CASCADE;
DROP TABLE products CASCADE;
DROP TABLE recommendations CASCADE;
DROP TABLE statistics CASCADE;
DROP TABLE purchases CASCADE;
DROP TABLE likes CASCADE;
DROP TABLE dislikes CASCADE;
DROP TABLE user_category_preferences CASCADE;
DROP TABLE product_analytics CASCADE;
DROP TABLE user_analytics CASCADE;