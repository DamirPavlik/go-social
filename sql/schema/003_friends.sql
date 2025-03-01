-- +goose Up
CREATE TABLE friends (
    id SERIAL PRIMARY KEY,
    user1 INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2 INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT now(),
    UNIQUE (user1, user2) 
);

-- +goose Down
DROP TABLE friends 