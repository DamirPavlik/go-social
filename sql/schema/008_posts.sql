-- +goose Up
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    image TEXT DEFAULT NULL, -- Optional post image
    created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE posts;
